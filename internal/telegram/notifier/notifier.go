package notifier

import (
	"log/slog"

	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/layout"
)

type Notifier struct {
	bot    *tele.Bot
	layout *layout.Layout
}

func New(bot *tele.Bot, layout *layout.Layout) *Notifier {
	return &Notifier{
		bot:    bot,
		layout: layout,
	}
}

func (n *Notifier) Notify(userID int64, textKey string, data any) error {
	recipient := &tele.User{ID: userID}

	text := n.layout.TextLocale("ru", textKey, data)
	markup := n.layout.MarkupLocale("ru", "core:hide")

	_, err := n.bot.Send(recipient, text, markup)
	if err != nil {
		slog.Error("failed to send notification", slog.Int64("user_id", userID), slog.String("error", err.Error()))
		return err
	}

	return nil
}
