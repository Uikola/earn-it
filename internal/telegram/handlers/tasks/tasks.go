package tasks

import (
	"context"
	"log/slog"
	"time"

	tele "gopkg.in/telebot.v3"
)

func (h *Handler) Tasks(c tele.Context) error {
	return h.showTasks(c, "today")
}

func (h *Handler) Today(c tele.Context) error {
	return h.showTasks(c, "today")
}

func (h *Handler) Tomorrow(c tele.Context) error {
	return h.showTasks(c, "tomorrow")
}

func (h *Handler) Later(c tele.Context) error {
	return h.showTasks(c, "later")
}

func (h *Handler) showTasks(c tele.Context, view string) error {
	ctx := context.Background()
	userID := c.Sender().ID

	user := h.userByIDWithProcessedError(ctx, c, userID)
	if user == nil {
		return nil
	}

	loc, err := time.LoadLocation(user.Timezone)
	if err != nil {
		loc = time.UTC
	}

	now := time.Now().In(loc)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

	if view == "today" {
		if err := h.taskRepository.RescheduleExpiredTasks(ctx, userID, today); err != nil {
			slog.Error("failed RescheduleExpiredTasks for today", slog.String("err", err.Error()))
			return c.Edit(
				h.layout.Text(c, "technical_issues"),
				h.layout.Markup(c, "mainMenuBack"),
			)
		}
	}

	var from, to time.Time
	var textKey string

	switch view {
	case "today":
		from = today
		to = today.AddDate(0, 0, 1)
		textKey = "tasks_menu_today"
	case "tomorrow":
		from = today.AddDate(0, 0, 1)
		to = today.AddDate(0, 0, 2)
		textKey = "tasks_menu_tomorrow"
	case "later":
		from = today.AddDate(0, 0, 2)
		to = today.AddDate(0, 10, 0)
		textKey = "tasks_menu_later"
	}

	tasks, err := h.taskRepository.TasksByUserAndDateRange(ctx, userID, from, to)
	if err != nil {
		slog.Error(
			"failed TasksByUserAndDateRange",
			slog.String("err", err.Error()),
			slog.Int64("userID", userID),
			slog.Time("from", from),
			slog.Time("to", to),
		)
		return c.Edit(
			h.layout.Text(c, "technical_issues"),
			h.layout.Markup(c, "mainMenuBack"),
		)
	}

	tasksToPrint := h.tasksToPrint(user.Timezone, tasks)

	var markup *tele.ReplyMarkup
	switch view {
	case "today":
		markup = h.layout.Markup(c, "tasksMenuToday")
	case "tomorrow":
		markup = h.layout.Markup(c, "tasksMenuTomorrow")
	default:
		markup = h.layout.Markup(c, "tasksMenuLater")
	}

	if len(tasksToPrint) == 0 {
		return c.Edit(
			h.layout.Text(c, "tasks_empty"),
			markup,
		)
	}

	return c.Edit(
		h.layout.Text(c, textKey, struct {
			Tasks []taskToPrint
			Count int
		}{
			Tasks: tasksToPrint,
			Count: len(tasksToPrint),
		}),
		markup,
	)
}
