package tasks

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	tele "gopkg.in/telebot.v3"
)

func (h *Handler) CompleteTasks(c tele.Context) error {
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

	if err := h.taskRepository.RescheduleExpiredTasks(ctx, userID, today); err != nil {
		slog.Error("failed to reschedule expired tasks", slog.String("err", err.Error()))
		return c.Edit(
			h.layout.Text(c, "technical_issues"),
			h.layout.Markup(c, "mainMenuBack"),
		)
	}

	tasks, err := h.taskRepository.TasksByUserAndDateRange(ctx, userID, today, today.AddDate(0, 0, 1))
	if err != nil {
		slog.Error("failed get tasks by user and date range", slog.String("err", err.Error()))
		return c.Edit(
			h.layout.Text(c, "technical_issues"),
			h.layout.Markup(c, "mainMenuBack"),
		)
	}

	var rows []tele.Row
	markup := c.Bot().NewMarkup()
	for _, task := range tasks {
		rows = append(rows, markup.Row(*h.layout.Button(c, "tasks:complete:task", struct {
			ID    int64
			Title string
		}{
			ID:    task.ID,
			Title: task.Title,
		})))
	}

	rows = append(rows, markup.Row(*h.layout.Button(c, "tasksMenuBack")))
	markup.Inline(rows...)

	return c.Edit(
		h.layout.Text(c, "task_complete_text"),
		markup,
	)
}

func (h *Handler) CompleteTask(c tele.Context) error {
	ctx := context.Background()
	userID := c.Sender().ID

	user := h.userByIDWithProcessedError(ctx, c, userID)
	if user == nil {
		return nil
	}

	taskID, err := strconv.ParseInt(c.Callback().Data, 10, 64)
	if err != nil {
		slog.Error("failed parse task id from callback", slog.String("err", err.Error()))
		return c.Edit(
			h.layout.Text(c, "technical_issues"),
			h.layout.Markup(c, "mainMenuBack"),
		)
	}

	task, err := h.taskRepository.TaskByID(ctx, taskID)
	if err != nil {
		slog.Error("failed get task by id", slog.String("err", err.Error()))
		return c.Edit(
			h.layout.Text(c, "technical_issues"),
			h.layout.Markup(c, "mainMenuBack"),
		)
	}

	err = h.transactor.WithinTransaction(ctx, func(txctx context.Context) error {
		now := time.Now()
		task.CompletedAt = &now
		task.Status = "done"
		if err := h.taskRepository.UpdateTask(txctx, taskID, task); err != nil {
			return fmt.Errorf("failed to complete task: %w", err)
		}

		if _, err := h.transactionRepository.CreateTransaction(txctx, userID, task.RewardValue, "task", task.ID); err != nil {
			return fmt.Errorf("failed to create transaction: %w", err)
		}

		user.Balance += task.RewardValue
		if err := h.userRepository.UpdateUser(txctx, *user); err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}

		return nil
	})
	if err != nil {
		slog.Error("failed complete task", slog.String("err", err.Error()))
		return c.Edit(
			h.layout.Text(c, "technical_issues"),
			h.layout.Markup(c, "mainMenuBack"),
		)
	}

	loc, err := time.LoadLocation(user.Timezone)
	if err != nil {
		loc = time.UTC
	}

	now := time.Now().In(loc)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	tomorrow := today.AddDate(0, 0, 1)

	tasks, err := h.taskRepository.TasksByUserAndDateRange(ctx, userID, today, tomorrow)
	if err != nil {
		slog.Error("failed get tasks by user", slog.String("err", err.Error()))
		return c.Edit(
			h.layout.Text(c, "technical_issues"),
			h.layout.Markup(c, "mainMenuBack"),
		)
	}

	tasksToPrint := h.tasksToPrint(user.Timezone, tasks)

	if len(tasksToPrint) == 0 {
		return c.Edit(
			h.layout.Text(c, "tasks_empty"),
			h.layout.Markup(c, "tasksMenuToday"),
		)
	}

	return c.Edit(
		h.layout.Text(c, "tasks_menu_today", struct {
			Tasks []taskToPrint
			Count int
		}{
			Tasks: tasksToPrint,
			Count: len(tasksToPrint),
		}),
		h.layout.Markup(c, "tasksMenuToday"),
	)
}
