package members

import "context"

// Notifier пишет участнику в Telegram; nil, если бот выключен.
type Notifier interface {
	SendMessage(ctx context.Context, chatID int64, text string) error
}
