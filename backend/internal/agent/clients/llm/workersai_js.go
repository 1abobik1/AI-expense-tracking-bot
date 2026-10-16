//go:build js && wasm

package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"syscall/js"

	"expensebot/internal/agent"

	"github.com/syumai/workers-go/cloudflare"
)

// workersAI вызывает env.AI.run — binding без ключей.
type workersAI struct {
	binding js.Value
	cfg     Config
}

func newWorkersAI(cfg Config) (agent.LLM, error) {
	binding := cloudflare.GetBinding("AI")
	if binding.IsUndefined() {
		return nil, errors.New("llm workersai: binding AI is not configured in wrangler")
	}

	return &workersAI{binding: binding, cfg: cfg}, nil
}

func (w *workersAI) Complete(ctx context.Context, req agent.LLMRequest) (string, error) {
	body := chatBody(w.cfg, req)
	delete(body, "model")

	input, err := toJS(body)
	if err != nil {
		return "", err
	}

	result, err := await(ctx, w.binding.Call("run", w.cfg.Model, input))
	if err != nil {
		return "", fmt.Errorf("llm workersai: %w", err)
	}

	raw := js.Global().Get("JSON").Call("stringify", result).String()
	// Ответ — {"response": строка или JSON} или OpenAI-подобный {"choices": [...]}.
	var out struct {
		Response json.RawMessage `json:"response"`
		Choices  []struct {
			Message chatMessage `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return "", fmt.Errorf("llm workersai: decode response: %w", err)
	}

	switch {
	case len(out.Choices) > 0:
		return out.Choices[0].Message.Content, nil
	case len(out.Response) > 0:
		var s string
		if json.Unmarshal(out.Response, &s) == nil {
			return s, nil
		}

		return string(out.Response), nil
	}

	return "", errors.New("llm workersai: empty response")
}

func toJS(v any) (js.Value, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return js.Value{}, err
	}

	return js.Global().Get("JSON").Call("parse", string(b)), nil
}

// await освобождает колбэки только после завершения Promise: вызов освобождённой функции роняет рантайм.
func await(ctx context.Context, promise js.Value) (js.Value, error) {
	type result struct {
		value js.Value
		err   error
	}
	done := make(chan result, 1)
	var onResolve, onReject js.Func
	settle := func(r result) {
		done <- r

		onResolve.Release()
		onReject.Release()
	}
	onResolve = js.FuncOf(func(_ js.Value, args []js.Value) any {
		settle(result{value: args[0]})
		return nil
	})
	onReject = js.FuncOf(func(_ js.Value, args []js.Value) any {
		settle(result{err: errors.New(args[0].Call("toString").String())})
		return nil
	})
	promise.Call("then", onResolve, onReject)

	select {
	case r := <-done:
		return r.value, r.err
	case <-ctx.Done():
		return js.Value{}, ctx.Err()
	}
}
