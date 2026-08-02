package shop

import (
	"context"
	"log/slog"

	"github.com/nlypage/intele"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/layout"

	"github.com/Uikola/earn-it/internal/models"
	"github.com/Uikola/earn-it/internal/repository/postgres"
)

type shopRepository interface {
	ShopItemsByUserID(ctx context.Context, userID int64, isPurchased bool) ([]models.ShopItem, error)
	ShopItemByID(ctx context.Context, id int64) (models.ShopItem, error)
	CreateShopItem(ctx context.Context, userID int64, name string, price int32) (models.ShopItem, error)
	MarkShopItemPurchased(ctx context.Context, id int64) error
	CreatePurchase(ctx context.Context, userID, shopItemID int64, pricePaid int32) (models.Purchase, error)
}

type userRepository interface {
	UserByID(ctx context.Context, id int64) (models.User, error)
	UpdateUser(ctx context.Context, user models.User) error
}

type transactionRepository interface {
	CreateTransaction(ctx context.Context, userID int64, amount int32, source, sourceName string) (models.Transaction, error)
}

type Handler struct {
	layout                *layout.Layout
	input                 *intele.InputManager
	transactor            postgres.Transactor
	shopRepository        shopRepository
	userRepository        userRepository
	transactionRepository transactionRepository
}

func NewHandler(
	layout *layout.Layout,
	input *intele.InputManager,
	transactor postgres.Transactor,
	shopRepository shopRepository,
	userRepository userRepository,
	transactionRepository transactionRepository,
) *Handler {
	return &Handler{
		layout:                layout,
		input:                 input,
		transactor:            transactor,
		shopRepository:        shopRepository,
		userRepository:        userRepository,
		transactionRepository: transactionRepository,
	}
}

type shopItemToPrint struct {
	Number int
	Name   string
	Price  int32
}

func (h *Handler) userByIDWithProcessedError(ctx context.Context, c tele.Context, userID int64, action string) *models.User {
	user, err := h.userRepository.UserByID(ctx, userID)
	if err != nil {
		slog.Error("failed to fetch user", slog.String("err", err.Error()), slog.Int64("userID", userID))

		switch action {
		case "send":
			if err := c.Send(
				h.layout.Text(c, "technical_issues"),
				h.layout.Markup(c, "mainMenuBack"),
			); err != nil {
				slog.Error("send error while userByIDWithProcessedError", slog.String("err", err.Error()))
				return nil
			}
		case "edit":
			if err := c.Edit(
				h.layout.Text(c, "technical_issues"),
				h.layout.Markup(c, "mainMenuBack"),
			); err != nil {
				slog.Error("edit error while userByIDWithProcessedError", slog.String("err", err.Error()))
				return nil
			}
		}

		return nil
	}

	return &user
}

func (h *Handler) shopItemsToPrint(items []models.ShopItem) []shopItemToPrint {
	result := make([]shopItemToPrint, 0, len(items))
	for i, item := range items {
		result = append(result, shopItemToPrint{
			Number: i + 1,
			Name:   item.Name,
			Price:  item.Price,
		})
	}
	return result
}
