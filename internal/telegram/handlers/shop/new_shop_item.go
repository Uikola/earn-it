package shop

import (
	"context"
	"errors"
	"log/slog"
	"strconv"

	tele "gopkg.in/telebot.v3"

	"github.com/Uikola/earn-it/internal/telegram/handlers/helpers"
)

func (h *Handler) NewShopItem(c tele.Context) error {
	ctx := context.Background()
	userID := c.Sender().ID

	steps := []helpers.InputStep{
		{
			Name:      "shop_item_name",
			PromptKey: "input_shop_item_name",
			Validator: validateShopItemName,
			ErrorKey:  "invalid_shop_item_name",
		},
		{
			Name:      "shop_item_price",
			PromptKey: "input_shop_item_price",
			Validator: validateShopItemPrice,
			ErrorKey:  "invalid_shop_item_price",
		},
	}

	results, err := helpers.CollectInput(c, h.input, h.layout, steps, h.layout.Markup(c, "shopMenuBack"), false)
	if err != nil {
		if errors.Is(err, helpers.ErrCanceled) {
			return nil
		}
		slog.Error("failed to collect input for new shop item", slog.String("err", err.Error()))
		return c.Send(
			h.layout.Text(c, "technical_issues"),
			h.layout.Markup(c, "mainMenuBack"),
		)
	}

	shopItemName := results["shop_item_name"]
	shopItemPriceTemp, _ := strconv.ParseInt(results["shop_item_price"], 10, 32)
	shopItemPrice := int32(shopItemPriceTemp)

	_, err = h.shopRepository.CreateShopItem(ctx, userID, shopItemName, shopItemPrice)
	if err != nil {
		slog.Error("failed to create shop item", slog.String("err", err.Error()), slog.Int64("userID", userID))
		return c.Send(
			h.layout.Text(c, "technical_issues"),
			h.layout.Markup(c, "mainMenuBack"),
		)
	}

	return h.ShopSend(c)
}
