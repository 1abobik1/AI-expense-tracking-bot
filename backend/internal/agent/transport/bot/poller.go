package bot

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"expensebot/internal/app/telegram"
)

const (
	pollTimeoutSec = 30
	// Сообщения одного пользователя идут в один воркер, чтобы не поменяться местами.
	workers = 8
)

type Poller struct {
	tg      *telegram.Client
	handler *Handler
	log     *slog.Logger
}

func NewPoller(tg *telegram.Client, handler *Handler, log *slog.Logger) *Poller {
	return &Poller{tg: tg, handler: handler, log: log}
}

// Run дожидается обработки уже полученных апдейтов после отмены ctx.
func (p *Poller) Run(ctx context.Context) error {
	if err := p.tg.DeleteWebhook(ctx); err != nil {
		return err
	}

	queues := make([]chan telegram.Update, workers)

	var wg sync.WaitGroup

	for i := range queues {
		queues[i] = make(chan telegram.Update, 16)
		wg.Go(func() {
			for u := range queues[i] {
				p.handler.HandleUpdate(context.WithoutCancel(ctx), u)
			}
		})
	}
	defer func() {
		for _, q := range queues {
			close(q)
		}

		wg.Wait()
	}()

	p.log.Info("telegram polling started")

	var offset int64
	for ctx.Err() == nil {
		updates, err := p.tg.GetUpdates(ctx, offset, pollTimeoutSec)
		if err != nil {
			if ctx.Err() == nil {
				p.log.Error("get updates", "err", err)
				sleep(ctx, 3*time.Second)
			}

			continue
		}

		for _, u := range updates {
			offset = u.UpdateID + 1
			queues[queueFor(u)] <- u
		}
	}

	return nil
}

func queueFor(u telegram.Update) int {
	if u.Message == nil || u.Message.From == nil {
		return 0
	}

	id := u.Message.From.ID
	if id < 0 {
		id = -id
	}

	return int(id % workers)
}

func sleep(ctx context.Context, d time.Duration) {
	select {
	case <-ctx.Done():
	case <-time.After(d):
	}
}
