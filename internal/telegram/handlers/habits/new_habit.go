package habits

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"time"

	tele "gopkg.in/telebot.v3"

	"github.com/Uikola/earn-it/internal/telegram/handlers/helpers"
	"github.com/Uikola/earn-it/internal/timeutil"
)

func (h *Handler) NewHabit(c tele.Context) error {
	ctx := context.Background()
	userID := c.Sender().ID

	steps := []helpers.InputStep{
		{
			Name:      "habit_name",
			PromptKey: "input_habit_name",
			Validator: validateHabitName,
			ErrorKey:  "invalid_habit_name",
		},

		{
			Name:      "habit_weekly_goal",
			PromptKey: "input_habit_weekly_goal",
			Validator: validateNumber,
			ErrorKey:  "invalid_habit_weekly_goal",
		},
		{
			Name:      "habit_reward_per_execute",
			PromptKey: "input_habit_reward_per_execute",
			Validator: validateNumber,
			ErrorKey:  "invalid_habit_reward_per_execute",
		},
	}

	results, err := helpers.CollectInput(c, h.input, h.layout, steps, h.layout.Markup(c, "habitsMenuBack"), false)
	if err != nil {
		if errors.Is(err, helpers.ErrCanceled) {
			return nil
		}
		slog.Error("failed to collect input", slog.String("err", err.Error()))
		return c.Send(h.layout.Text(c, "technical_issues"), h.layout.Markup(c, "mainMenuBack"))
	}

	var (
		habitName             string
		habitWeeklyGoal       int32
		habitRewardForExecute int32
	)

	habitName = results["habit_name"]

	habitWeeklyGoalTemp, _ := strconv.ParseInt(results["habit_weekly_goal"], 10, 32)
	habitWeeklyGoal = int32(habitWeeklyGoalTemp)

	habitRewardForExecuteTemp, _ := strconv.ParseInt(results["habit_reward_per_execute"], 10, 32)
	habitRewardForExecute = int32(habitRewardForExecuteTemp)

	_, err = h.habitsRepository.CreateHabit(ctx, userID, habitName, habitWeeklyGoal, habitRewardForExecute)
	if err != nil {
		slog.Error("failed to create habit", slog.String("err", err.Error()), slog.Int64("userID", userID))
		return c.Send(
			h.layout.Text(c, "technical_issues"),
			h.layout.Markup(c, "mainMenuBack"),
		)
	}

	user := h.userByIDWithProcessedError(ctx, c, userID, "send")
	if user == nil {
		return nil
	}

	habits := h.habitsByUserIDWithProcessedError(ctx, c, userID, "send")
	if habits == nil {
		return nil
	}

	loc, err := time.LoadLocation(user.Timezone)
	if err != nil {
		loc = time.UTC
	}
	weekStart := timeutil.WeekStart(time.Now(), loc)

	return c.Send(
		h.layout.Text(c, "habits_menu_text", struct {
			Habits            []habitToPrint
			RewardWeeklyBonus int32
		}{
			Habits:            h.habitsToPrint(ctx, c, weekStart, habits),
			RewardWeeklyBonus: user.RewardWeeklyBonus,
		}),
		h.layout.Markup(c, "habitsMenu"),
	)
}
