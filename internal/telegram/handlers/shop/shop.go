package shop

import (
	"context"
	"log/slog"

	tele "gopkg.in/telebot.v3"
)

func (h *Handler) Shop(c tele.Context) error {
	ctx := context.Background()
	userID := c.Sender().ID

	user := h.userByIDWithProcessedError(ctx, c, userID, "edit")
	if user == nil {
		return nil
	}

	items, err := h.shopRepository.ShopItemsByUserID(ctx, userID, false)
	if err != nil {
		slog.Error("failed to fetch shop items", slog.String("err", err.Error()), slog.Int64("userID", userID))
		return c.Edit(
			h.layout.Text(c, "technical_issues"),
			h.layout.Markup(c, "mainMenuBack"),
		)
	}

	itemsToPrint := h.shopItemsToPrint(items)

	return c.Edit(
		h.layout.Text(c, "shop_menu_text", struct {
			Items   []shopItemToPrint
			Balance int32
		}{
			Items:   itemsToPrint,
			Balance: user.Balance,
		}),
		h.layout.Markup(c, "shopMenu"),
	)
}

func (h *Handler) ShopSend(c tele.Context) error {
	ctx := context.Background()
	userID := c.Sender().ID

	user, err := h.userRepository.UserByID(ctx, userID)
	if err != nil {
		slog.Error("failed to fetch user", slog.String("err", err.Error()), slog.Int64("userID", userID))
		return c.Send(
			h.layout.Text(c, "technical_issues"),
			h.layout.Markup(c, "mainMenuBack"),
		)
	}

	items, err := h.shopRepository.ShopItemsByUserID(ctx, userID, false)
	if err != nil {
		slog.Error("failed to fetch shop items", slog.String("err", err.Error()), slog.Int64("userID", userID))
		return c.Send(
			h.layout.Text(c, "technical_issues"),
			h.layout.Markup(c, "mainMenuBack"),
		)
	}

	itemsToPrint := h.shopItemsToPrint(items)

	return c.Send(
		h.layout.Text(c, "shop_menu_text", struct {
			Items   []shopItemToPrint
			Balance int32
		}{
			Items:   itemsToPrint,
			Balance: user.Balance,
		}),
		h.layout.Markup(c, "shopMenu"),
	)
}
