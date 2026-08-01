package habits

import (
	"context"
	"errors"
	"strconv"

	"github.com/Uikola/earn-it/internal/telegram/handlers/helpers"
	"github.com/google/martian/log"
	tele "gopkg.in/telebot.v3"
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
			Name:      "habit_weakly_goal",
			PromptKey: "input_habit_weakly_goal",
			Validator: validateNumber,
			ErrorKey:  "invalid_habit_weakly_goal",
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
		log.Errorf("failed to collect input: %v", err)
		return c.Send(h.layout.Text(c, "technical_issues"), h.layout.Markup(c, "mainMenuBack"))
	}

	var (
		habitName             string
		habitWeaklyGoal       int32
		habitRewardForExecute int32
	)

	habitName = results["habit_name"]

	habitWeaklyGoalTemp, _ := strconv.ParseInt(results["habit_weakly_goal"], 10, 32)
	habitWeaklyGoal = int32(habitWeaklyGoalTemp)

	habitRewardForExecuteTemp, _ := strconv.ParseInt(results["habit_reward_per_execute"], 10, 32)
	habitRewardForExecute = int32(habitRewardForExecuteTemp)

	_, err = h.habitsRepository.CreateHabit(ctx, userID, habitName, habitWeaklyGoal, habitRewardForExecute)
	if err != nil {
		log.Errorf("failed to create habit: %v", err)
		return c.Send(
			h.layout.Text(c, "technical_issues"),
			h.layout.Markup(c, "mainMenuBack"),
		)
	}

	user := h.userByIDWithProcessedError(ctx, c, userID, "send")
	if user == nil {
		return nil
	}

	habits := h.habitsByUserIDWithProcessedError(ctx, c, userID, "edit")
	if habits == nil {
		return nil
	}

	return c.Send(
		h.layout.Text(c, "habits_menu_text", struct {
			Habits            []habitToPrint
			RewardWeeklyBonus int32
		}{
			Habits:            h.habitsToPrint(ctx, c, user.Timezone, habits),
			RewardWeeklyBonus: user.RewardWeeklyBonus,
		}),
		h.layout.Markup(c, "habitsMenu"),
	)
}
