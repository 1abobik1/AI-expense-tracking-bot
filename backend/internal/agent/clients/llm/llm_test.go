package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"expensebot/internal/agent"
)

func TestOpenAICompatible(t *testing.T) {
	var got map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" || r.Header.Get("Authorization") != "Bearer key" {
			t.Errorf("request %s, auth %q", r.URL.Path, r.Header.Get("Authorization"))
		}

		json.NewDecoder(r.Body).Decode(&got)
		w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"{\"action\":\"add\"}"}}]}`))
	}))
	defer srv.Close()

	client, err := New(Config{
		Provider: "openai", Model: "@cf/google/gemma", BaseURL: srv.URL + "/v1/", APIKey: "key",
		ExtraBody: map[string]any{"chat_template_kwargs": map[string]any{"enable_thinking": false}},
	}, srv.Client())
	if err != nil {
		t.Fatal(err)
	}

	text, err := client.Complete(context.Background(), agent.LLMRequest{System: "sys", User: "кофе 350", MaxTokens: 100})
	if err != nil {
		t.Fatal(err)
	}

	if text != `{"action":"add"}` {
		t.Fatalf("text = %q", text)
	}

	if got["model"] != "@cf/google/gemma" || got["chat_template_kwargs"] == nil || len(got["messages"].([]any)) != 2 {
		t.Fatalf("request body = %v", got)
	}
}

func TestOpenAIErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"quota"}`, http.StatusTooManyRequests)
	}))
	defer srv.Close()

	client, _ := New(Config{Provider: "openai", Model: "m", BaseURL: srv.URL}, srv.Client())
	if _, err := client.Complete(context.Background(), agent.LLMRequest{}); err == nil {
		t.Fatal("want error on HTTP 429")
	}
}

func TestAnthropic(t *testing.T) {
	var got map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/messages" || r.Header.Get("X-Api-Key") != "key" {
			t.Errorf("request %s, key %q", r.URL.Path, r.Header.Get("X-Api-Key"))
		}

		json.NewDecoder(r.Body).Decode(&got)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"msg_1","type":"message","role":"assistant","model":"claude-opus-5-5",
			"content":[{"type":"text","text":"{\"action\":\"list\"}"}],"stop_reason":"end_turn",
			"usage":{"input_tokens":1,"output_tokens":1}}`))
	}))
	defer srv.Close()

	client, err := New(Config{Provider: "anthropic", Model: "claude-opus-5-5", BaseURL: srv.URL, APIKey: "key"}, srv.Client())
	if err != nil {
		t.Fatal(err)
	}

	text, err := client.Complete(context.Background(), agent.LLMRequest{System: "sys", User: "покажи траты", MaxTokens: 100})
	if err != nil {
		t.Fatal(err)
	}

	if text != `{"action":"list"}` {
		t.Fatalf("text = %q", text)
	}

	if got["model"] != "claude-opus-5-5" || got["temperature"] != nil {
		t.Fatalf("request body = %v", got)
	}
}

func TestOpenRouterDefaultsBaseURL(t *testing.T) {
	c, err := New(Config{Provider: "openrouter", Model: "deepseek/deepseek-v4-flash", APIKey: "k"}, http.DefaultClient)
	if err != nil {
		t.Fatal(err)
	}

	if got := c.(*openAI).cfg.BaseURL; got != openRouterBaseURL {
		t.Fatalf("base URL = %q", got)
	}
}

func TestUnknownProvider(t *testing.T) {
	if _, err := New(Config{Provider: "nope", Model: "m"}, http.DefaultClient); err == nil {
		t.Fatal("want error")
	}

	if _, err := New(Config{Provider: "workersai", Model: "m"}, http.DefaultClient); err == nil {
		t.Fatal("workersai outside Worker must fail")
	}
}
