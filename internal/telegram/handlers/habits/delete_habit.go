package habits

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	tele "gopkg.in/telebot.v3"

	"github.com/Uikola/earn-it/internal/timeutil"
)

func (h *Handler) DeleteHabits(c tele.Context) error {
	ctx := context.Background()
	userID := c.Sender().ID

	habits := h.habitsByUserIDWithProcessedError(ctx, c, userID, "edit")
	if habits == nil {
		return nil
	}

	var rows []tele.Row

	markup := c.Bot().NewMarkup()
	for _, habit := range habits {
		rows = append(rows, markup.Row(*h.layout.Button(c, "habits:delete:habit", struct {
			ID   int64
			Name string
		}{
			ID:   habit.ID,
			Name: habit.Name,
		})))
	}

	rows = append(
		rows,
		markup.Row(*h.layout.Button(c, "habitsMenuBack")),
	)
	markup.Inline(rows...)

	return c.Edit(
		h.layout.Text(c, "habit_delete_text"),
		markup,
	)
}

func (h *Handler) DeleteHabit(c tele.Context) error {
	ctx := context.Background()
	userID := c.Sender().ID

	habitID, err := strconv.ParseInt(c.Callback().Data, 10, 64)
	if err != nil {
		slog.Error("invalid callback data for complete habit handler", slog.String("err", err.Error()))
		return c.Edit(
			h.layout.Text(c, "technical_issues"),
			h.layout.Markup(c, "mainMenuBack"),
		)
	}

	if err := h.habitsRepository.DeleteHabit(ctx, habitID); err != nil {
		slog.Error("failed to delete habit log", slog.String("err", err.Error()), slog.Int64("habitID", habitID))
		return c.Edit(
			h.layout.Text(c, "technical_issues"),
			h.layout.Markup(c, "mainMenuBack"),
		)
	}

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
