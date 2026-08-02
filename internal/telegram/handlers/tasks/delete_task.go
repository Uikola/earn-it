package tasks

import (
	"context"
	"log/slog"
	"strconv"

	tele "gopkg.in/telebot.v3"
)

func (h *Handler) DeleteTasks(c tele.Context) error {
	ctx := context.Background()
	userID := c.Sender().ID

	tasks, err := h.taskRepository.TasksByUserID(ctx, userID)
	if err != nil {
		slog.Error("failed TasksByUserID", slog.String("err", err.Error()), slog.Int64("userID", userID))
		return c.Edit(
			h.layout.Text(c, "technical_issues"),
			h.layout.Markup(c, "mainMenuBack"),
		)
	}

	var rows []tele.Row

	markup := c.Bot().NewMarkup()
	for _, task := range tasks {
		rows = append(rows, markup.Row(*h.layout.Button(c, "tasks:delete:task", struct {
			ID    int64
			Title string
		}{
			ID:    task.ID,
			Title: task.Title,
		})))
	}

	rows = append(
		rows,
		markup.Row(*h.layout.Button(c, "tasksMenuBack")),
	)
	markup.Inline(rows...)

	return c.Edit(
		h.layout.Text(c, "task_delete_text"),
		markup,
	)
}

func (h *Handler) DeleteTask(c tele.Context) error {
	ctx := context.Background()

	taskID, err := strconv.ParseInt(c.Callback().Data, 10, 64)
	if err != nil {
		slog.Error("invalid callback data for delete task handler", slog.String("err", err.Error()))
		return c.Edit(
			h.layout.Text(c, "technical_issues"),
			h.layout.Markup(c, "mainMenuBack"),
		)
	}

	if err := h.taskRepository.DeleteTask(ctx, taskID); err != nil {
		slog.Error("failed to delete task", slog.String("err", err.Error()), slog.Int64("taskID", taskID))
		return c.Edit(
			h.layout.Text(c, "technical_issues"),
			h.layout.Markup(c, "mainMenuBack"),
		)
	}

	return h.showTasks(c, "today")
}
