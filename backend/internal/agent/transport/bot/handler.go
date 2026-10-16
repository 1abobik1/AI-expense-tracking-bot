// Package bot — транспорт Telegram: апдейт → проверки → агент → ответ.
package bot

import (
	"context"
	"log/slog"
	"strings"

	"expensebot/internal/agent"
	"expensebot/internal/agent/usecase"
	"expensebot/internal/app/telegram"
	"expensebot/internal/members"

	membersModels "expensebot/internal/members/models"
)

type Handler struct {
	agentUC   agent.AgentUseCase
	membersUC members.MembersUseCase
	tg        agent.Messenger
	log       *slog.Logger
}

func NewHandler(agentUC agent.AgentUseCase, membersUC members.MembersUseCase, tg agent.Messenger, log *slog.Logger) *Handler {
	return &Handler{agentUC: agentUC, membersUC: membersUC, tg: tg, log: log}
}

const (
	pendingText  = "Заявка отправлена главе семьи. Как только доступ откроют, я напишу."
	rejectedText = "Доступ к боту закрыт."
	textOnly     = "Я понимаю только текст. Напиши, например: «кофе 350»."
	somethingBad = "Что-то пошло не так, попробуй ещё раз."
)

// HandleUpdate не возвращает ошибок: повторная доставка апдейта записала бы трату дважды.
func (h *Handler) HandleUpdate(ctx context.Context, u telegram.Update) {
	msg := u.Message
	// Группы игнорируются: в общем чате траты увидели бы все.
	if msg == nil || msg.From == nil || msg.From.IsBot || msg.Chat.Type != "private" {
		return
	}

	member, err := h.membersUC.TouchMember(ctx, mapUserToMember(*msg.From))
	if err != nil {
		h.log.Error("touch member", "user_id", msg.From.ID, "err", err)
		h.reply(ctx, msg.Chat.ID, somethingBad)

		return
	}
	// До одобрения сообщения не доходят до ИИ: чужие не тратят квоту.
	switch member.Status {
	case membersModels.StatusPending:
		h.reply(ctx, msg.Chat.ID, pendingText)
		return
	case membersModels.StatusRejected:
		h.reply(ctx, msg.Chat.ID, rejectedText)
		return
	}

	text := strings.TrimSpace(msg.Text)
	switch {
	case text == "":
		h.reply(ctx, msg.Chat.ID, textOnly)
		return
	case strings.HasPrefix(text, "/"): // /start, /help и любые команды
		h.reply(ctx, msg.Chat.ID, usecase.HelpText)
		return
	}

	_ = h.tg.SendTyping(ctx, msg.Chat.ID)
	// Владелец трат — отправитель из апдейта, а не что-то из текста.
	reply, err := h.agentUC.Handle(ctx, member.UserID, text)
	if err != nil {
		h.log.Error("handle message", "user_id", member.UserID, "err", err)

		reply = somethingBad
	}

	h.reply(ctx, msg.Chat.ID, reply)
}

func (h *Handler) reply(ctx context.Context, chatID int64, text string) {
	if err := h.tg.SendMessage(ctx, chatID, text); err != nil {
		h.log.Error("send message", "chat_id", chatID, "err", err)
	}
}
