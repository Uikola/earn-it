package habits

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	tele "gopkg.in/telebot.v3"

	"github.com/Uikola/earn-it/internal/timeutil"
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

	loc, err := time.LoadLocation(user.Timezone)
	if err != nil {
		loc = time.UTC
	}
	weekStart := timeutil.WeekStart(time.Now(), loc)

	var rows []tele.Row
	markup := c.Bot().NewMarkup()
	for _, habit := range habits {
		habitLogsForWeek, err := h.habitsRepository.HabitLogsForWeek(ctx, habit.ID, weekStart)
		if err != nil {
			slog.Error("failed to fetch habit logs for week", slog.String("err", err.Error()), slog.Int64("habitID", habit.ID))
			return c.Edit(
				h.layout.Text(c, "technical_issues"),
				h.layout.Markup(c, "mainMenuBack"),
			)
		}

		rows = append(rows, markup.Row(*h.layout.Button(c, "habits:complete:habit", struct {
			ID                  int64
			Name                string
			WeeklyGoalDone      int32
			WeeklyGoal          int32
			WeeklyGoalCompleted bool
		}{
			ID:                  habit.ID,
			Name:                habit.Name,
			WeeklyGoalDone:      int32(len(habitLogsForWeek)),
			WeeklyGoal:          habit.WeeklyGoal,
			WeeklyGoalCompleted: int32(len(habitLogsForWeek)) >= habit.WeeklyGoal,
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
		slog.Error("invalid callback data for complete habit handler", slog.String("err", err.Error()))
		return c.Edit(
			h.layout.Text(c, "technical_issues"),
			h.layout.Markup(c, "mainMenuBack"),
		)
	}

	habit, err := h.habitsRepository.HabitByID(ctx, habitID)
	if err != nil {
		slog.Error("failed to fetch habit", slog.String("err", err.Error()), slog.Int64("habitID", habitID))
		return c.Edit(
			h.layout.Text(c, "technical_issues"),
			h.layout.Markup(c, "mainMenuBack"),
		)
	}

	err = h.transactor.WithinTransaction(ctx, func(txctx context.Context) error {
		habitLog, err := h.habitsRepository.CreateHabitLog(txctx, habitID)
		if err != nil {
			return fmt.Errorf("failed to create habit: %w", err)
		}

		if _, err := h.transactionRepository.CreateTransaction(txctx, userID, habit.RewardPerExecute, "habit_log", habitLog.ID); err != nil {
			return fmt.Errorf("failed to create transaction: %w", err)
		}

		user.Balance += habit.RewardPerExecute
		if err := h.userRepository.UpdateUser(txctx, *user); err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}
		return nil
	})
	if err != nil {
		slog.Error("failed to complete transaction", slog.String("err", err.Error()), slog.Int64("habitID", habitID))
		return c.Edit(
			h.layout.Text(c, "technical_issues"),
			h.layout.Markup(c, "mainMenuBack"),
		)
	}

	loc, err := time.LoadLocation(user.Timezone)
	if err != nil {
		loc = time.UTC
	}
	weekStart := timeutil.WeekStart(time.Now(), loc)

	return c.Edit(
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
