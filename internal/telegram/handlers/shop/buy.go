package shop

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	tele "gopkg.in/telebot.v3"
)

func (h *Handler) BuyItems(c tele.Context) error {
	ctx := context.Background()
	userID := c.Sender().ID

	user := h.userByIDWithProcessedError(ctx, c, userID, "edit")
	if user == nil {
		return nil
	}

	items, err := h.shopRepository.ShopItemsByUserID(ctx, userID, false)
	if err != nil {
		slog.Error("failed to fetch shop items for buy", slog.String("err", err.Error()), slog.Int64("userID", userID))
		return c.Edit(
			h.layout.Text(c, "technical_issues"),
			h.layout.Markup(c, "mainMenuBack"),
		)
	}

	if len(items) == 0 {
		return c.Edit(
			h.layout.Text(c, "shop_empty"),
			h.layout.Markup(c, "shopMenu"),
		)
	}

	var rows []tele.Row
	markup := c.Bot().NewMarkup()
	for _, item := range items {
		rows = append(rows, markup.Row(*h.layout.Button(c, "shop:buy:item", struct {
			ID    int64
			Name  string
			Price int32
		}{
			ID:    item.ID,
			Name:  item.Name,
			Price: item.Price,
		})))
	}

	rows = append(rows, markup.Row(*h.layout.Button(c, "shopMenuBack")))
	markup.Inline(rows...)

	return c.Edit(
		h.layout.Text(c, "shop_buy_text"),
		markup,
	)
}

func (h *Handler) BuyItem(c tele.Context) error {
	ctx := context.Background()
	userID := c.Sender().ID

	itemID, err := strconv.ParseInt(c.Callback().Data, 10, 64)
	if err != nil {
		slog.Error("invalid callback data for buy shop item", slog.String("err", err.Error()))
		return c.Edit(
			h.layout.Text(c, "technical_issues"),
			h.layout.Markup(c, "mainMenuBack"),
		)
	}

	item, err := h.shopRepository.ShopItemByID(ctx, itemID)
	if err != nil {
		slog.Error("failed to fetch shop item by id", slog.String("err", err.Error()), slog.Int64("itemID", itemID))
		return c.Edit(
			h.layout.Text(c, "technical_issues"),
			h.layout.Markup(c, "mainMenuBack"),
		)
	}

	user := h.userByIDWithProcessedError(ctx, c, userID, "edit")
	if user == nil {
		return nil
	}

	if user.Balance < item.Price {
		return c.Respond(&tele.CallbackResponse{
			Text:      h.layout.Text(c, "shop_insufficient_funds"),
			ShowAlert: true,
		})
	}

	err = h.transactor.WithinTransaction(ctx, func(txctx context.Context) error {
		if _, err := h.transactionRepository.CreateTransaction(txctx, userID, -item.Price, "purchase", item.Name); err != nil {
			return fmt.Errorf("failed to create transaction: %w", err)
		}

		user.Balance -= item.Price
		if err := h.userRepository.UpdateUser(txctx, *user); err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}

		if err := h.shopRepository.MarkShopItemPurchased(txctx, itemID); err != nil {
			return fmt.Errorf("failed to mark item as purchased: %w", err)
		}

		if _, err := h.shopRepository.CreatePurchase(txctx, userID, itemID, item.Price); err != nil {
			return fmt.Errorf("failed to create purchase: %w", err)
		}

		return nil
	})
	if err != nil {
		slog.Error("failed to complete purchase transaction", slog.String("err", err.Error()), slog.Int64("itemID", itemID))
		return c.Edit(
			h.layout.Text(c, "technical_issues"),
			h.layout.Markup(c, "mainMenuBack"),
		)
	}

	return h.Shop(c)
}
