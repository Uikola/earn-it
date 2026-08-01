package habits

import (
	"context"

	tele "gopkg.in/telebot.v3"
)

func (h *Handler) Habits(c tele.Context) error {
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
