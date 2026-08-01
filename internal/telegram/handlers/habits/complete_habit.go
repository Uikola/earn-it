package habits

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/martian/log"
	tele "gopkg.in/telebot.v3"
)

func (h *Handler) CompleteHabits(c tele.Context) error {
	ctx := context.Background()
	userID := c.Sender().ID

	user := h.userByIDWithProcessedError(ctx, c, userID, "edit")
	if user == nil {
		return nil
	}

	habits := h.habitsByUserIDWithProcessedError(ctx, c, userID, "edit")
	if habits == nil {
		return nil
	}

	var rows []tele.Row
	markup := c.Bot().NewMarkup()
	for _, habit := range habits {
		habitLogsForWeek, err := h.habitsRepository.HabitLogsForWeek(ctx, habit.ID, user.Timezone)
		if err != nil {
			log.Errorf("failed to fetch habit logs for week: %v", err)
			return c.Edit(
				h.layout.Text(c, "technical_issues"),
				h.layout.Markup(c, "mainMenuBack"),
			)
		}

		rows = append(rows, markup.Row(*h.layout.Button(c, "habits:complete:habit", struct {
			ID                  int64
			Name                string
			WeaklyGoalDone      int32
			WeaklyGoal          int32
			WeaklyGoalCompleted bool
		}{
			ID:                  habit.ID,
			Name:                habit.Name,
			WeaklyGoalDone:      int32(len(habitLogsForWeek)),
			WeaklyGoal:          habit.WeaklyGoal,
			WeaklyGoalCompleted: int32(len(habitLogsForWeek)) >= habit.WeaklyGoal,
		})))
	}

	rows = append(
		rows,
		markup.Row(*h.layout.Button(c, "habitsMenuBack")),
	)
	markup.Inline(rows...)

	return c.Edit(
		h.layout.Text(c, "habit_complete_text"),
		markup,
	)
}

func (h *Handler) CompleteHabit(c tele.Context) error {
	ctx := context.Background()
	userID := c.Sender().ID

	user := h.userByIDWithProcessedError(ctx, c, userID, "edit")
	if user == nil {
		return nil
	}

	habits := h.habitsByUserIDWithProcessedError(ctx, c, userID, "edit")
	if habits == nil {
		return nil
	}

	habitID, err := strconv.ParseInt(c.Callback().Data, 10, 64)
	if err != nil {
		log.Errorf("invalid callback data for complete habit handler: %v", err)
		return c.Edit(
			h.layout.Text(c, "technical_issues"),
			h.layout.Markup(c, "mainMenuBack"),
		)
	}

	habit, err := h.habitsRepository.HabitByID(ctx, habitID)
	if err != nil {
		log.Errorf("failed to fetch habit: %v", err)
		return c.Edit(
			h.layout.Text(c, "technical_issues"),
			h.layout.Markup(c, "mainMenuBack"),
		)
	}

	err = h.transactor.WithinTransaction(ctx, func(txctx context.Context) error {
		if err := h.habitsRepository.CreateHabitLog(txctx, habitID); err != nil {
			return fmt.Errorf("failed to create habit: %v", err)
		}

		user.Balance += habit.RewardPerExecute
		if err := h.userRepository.UpdateUser(ctx, *user); err != nil {
			return fmt.Errorf("failed to update user: %v", err)
		}
		return nil
	})
	if err != nil {
		log.Errorf("failed to complete transaction: %v", err)
		return c.Edit(
			h.layout.Text(c, "technical_issues"),
			h.layout.Markup(c, "mainMenuBack"),
		)
	}

	return c.Edit(
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
