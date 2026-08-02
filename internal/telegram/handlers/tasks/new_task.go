package tasks

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"time"

	tele "gopkg.in/telebot.v3"

	"github.com/Uikola/earn-it/internal/telegram/handlers/helpers"
)

func (h *Handler) NewTask(c tele.Context) error {
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

	steps := []helpers.InputStep{
		{
			Name:      "task_title",
			PromptKey: "input_task_title",
			Validator: validateTaskTitle,
			ErrorKey:  "invalid_task_title",
		},
		{
			Name:      "task_date",
			PromptKey: "input_task_date",
			Validator: validateDate,
			ErrorKey:  "invalid_task_date",
		},
		{
			Name:      "task_reward",
			PromptKey: "input_task_reward",
			Validator: validateReward,
			ErrorKey:  "invalid_task_reward",
		},
	}

	results, err := helpers.CollectInput(c, h.input, h.layout, steps, h.layout.Markup(c, "tasksMenuBack"), false)
	if err != nil {
		slog.Error("failed collect input", slog.String("err", err.Error()))
		if errors.Is(err, helpers.ErrCanceled) {
			return nil
		}
		return c.Send(h.layout.Text(c, "technical_issues"), h.layout.Markup(c, "mainMenuBack"))
	}

	title := results["task_title"]

	scheduledDate, _ := time.ParseInLocation("2006-01-02", results["task_date"], loc)

	rewardValue := int32(10)
	if rewardStr, ok := results["task_reward"]; ok && rewardStr != "" {
		if reward, err := strconv.ParseInt(rewardStr, 10, 32); err == nil {
			rewardValue = int32(reward)
		}
	}

	_, err = h.taskRepository.CreateTask(ctx, userID, title, scheduledDate, rewardValue)
	if err != nil {
		slog.Error("failed create task", slog.String("err", err.Error()))
		return c.Send(
			h.layout.Text(c, "technical_issues"),
			h.layout.Markup(c, "mainMenuBack"),
		)
	}

	return c.Send(
		h.layout.Text(c, "task_created"),
		h.layout.Markup(c, "tasksMenuBack"),
	)
}
