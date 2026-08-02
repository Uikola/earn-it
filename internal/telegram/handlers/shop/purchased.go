package shop

import (
	"context"
	"log/slog"

	tele "gopkg.in/telebot.v3"
)

func (h *Handler) PurchasedItems(c tele.Context) error {
	ctx := context.Background()
	userID := c.Sender().ID

	items, err := h.shopRepository.ShopItemsByUserID(ctx, userID, true)
	if err != nil {
		slog.Error("failed to fetch purchased items", slog.String("err", err.Error()), slog.Int64("userID", userID))
		return c.Edit(
			h.layout.Text(c, "technical_issues"),
			h.layout.Markup(c, "mainMenuBack"),
		)
	}

	itemsToPrint := h.shopItemsToPrint(items)

	return c.Edit(
		h.layout.Text(c, "shop_purchased_text", struct {
			Items []shopItemToPrint
		}{
			Items: itemsToPrint,
		}),
		h.layout.Markup(c, "shopPurchasedBack"),
	)
}
