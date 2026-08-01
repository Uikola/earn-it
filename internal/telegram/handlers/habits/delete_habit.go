package habits

import (
	"context"
	"strconv"

	"github.com/google/martian/log"
	tele "gopkg.in/telebot.v3"
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
		log.Errorf("invalid callback data for complete habit handler: %v", err)
		return c.Edit(
			h.layout.Text(c, "technical_issues"),
			h.layout.Markup(c, "mainMenuBack"),
		)
	}

	if err := h.habitsRepository.DeleteHabit(ctx, habitID); err != nil {
		log.Errorf("failed to delete habit log: %v", err)
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
