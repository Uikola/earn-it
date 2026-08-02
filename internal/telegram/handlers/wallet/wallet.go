package wallet

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/layout"

	"github.com/Uikola/earn-it/internal/models"
)

type userRepository interface {
	UserByID(ctx context.Context, id int64) (models.User, error)
}

type transactionRepository interface {
	TotalIncome(ctx context.Context, userID int64) (int32, error)
	TotalExpense(ctx context.Context, userID int64) (int32, error)
	RecentTransactionsWithDetails(ctx context.Context, userID int64, limit int32) ([]models.TransactionWithDetails, error)
}

type Handler struct {
	layout                *layout.Layout
	userRepository        userRepository
	transactionRepository transactionRepository
}

func NewHandler(
	layout *layout.Layout,
	userRepository userRepository,
	transactionRepository transactionRepository,
) *Handler {
	return &Handler{
		layout:                layout,
		userRepository:        userRepository,
		transactionRepository: transactionRepository,
	}
}

type transactionToPrint struct {
	FormattedAmount string
	SourceName      string
	FormattedDate   string
}

func (h *Handler) Wallet(c tele.Context) error {
	ctx := context.Background()
	userID := c.Sender().ID

	user, err := h.userRepository.UserByID(ctx, userID)
	if err != nil {
		slog.Error("failed to fetch user", slog.String("err", err.Error()), slog.Int64("userID", userID))
		return c.Edit(
			h.layout.Text(c, "technical_issues"),
			h.layout.Markup(c, "mainMenuBack"),
		)
	}

	totalIncome, err := h.transactionRepository.TotalIncome(ctx, userID)
	if err != nil {
		slog.Error("failed to fetch total income", slog.String("err", err.Error()), slog.Int64("userID", userID))
		return c.Edit(
			h.layout.Text(c, "technical_issues"),
			h.layout.Markup(c, "mainMenuBack"),
		)
	}

	totalExpense, err := h.transactionRepository.TotalExpense(ctx, userID)
	if err != nil {
		slog.Error("failed to fetch total expense", slog.String("err", err.Error()), slog.Int64("userID", userID))
		return c.Edit(
			h.layout.Text(c, "technical_issues"),
			h.layout.Markup(c, "mainMenuBack"),
		)
	}

	recentTransactions, err := h.transactionRepository.RecentTransactionsWithDetails(ctx, userID, 5)
	if err != nil {
		slog.Error("failed to fetch recent transactions", slog.String("err", err.Error()), slog.Int64("userID", userID))
		return c.Edit(
			h.layout.Text(c, "technical_issues"),
			h.layout.Markup(c, "mainMenuBack"),
		)
	}

	transactionsToPrint := h.transactionsToPrint(user.Timezone, recentTransactions)

	return c.Edit(
		h.layout.Text(c, "wallet_menu_text", struct {
			Balance      int32
			TotalIncome  int32
			TotalExpense int32
			Transactions []transactionToPrint
		}{
			Balance:      user.Balance,
			TotalIncome:  totalIncome,
			TotalExpense: totalExpense,
			Transactions: transactionsToPrint,
		}),
		h.layout.Markup(c, "walletMenu"),
	)
}

func (h *Handler) transactionsToPrint(timezone string, transactions []models.TransactionWithDetails) []transactionToPrint {
	result := make([]transactionToPrint, 0, len(transactions))

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}

	now := time.Now().In(loc)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	yesterday := today.AddDate(0, 0, -1)

	for _, tx := range transactions {
		txTime := tx.CreatedAt.In(loc)
		txDate := time.Date(txTime.Year(), txTime.Month(), txTime.Day(), 0, 0, 0, 0, loc)

		var formattedAmount string
		if tx.Amount > 0 {
			formattedAmount = fmt.Sprintf("+%d", tx.Amount)
		} else {
			formattedAmount = fmt.Sprintf("%d", tx.Amount)
		}

		var formattedDate string
		switch {
		case txDate.Equal(today):
			formattedDate = fmt.Sprintf("сегодня %02d:%02d", txTime.Hour(), txTime.Minute())
		case txDate.Equal(yesterday):
			formattedDate = fmt.Sprintf("вчера %02d:%02d", txTime.Hour(), txTime.Minute())
		default:
			formattedDate = fmt.Sprintf("%02d.%02d %02d:%02d", txTime.Day(), txTime.Month(), txTime.Hour(), txTime.Minute())
		}

		result = append(result, transactionToPrint{
			FormattedAmount: formattedAmount,
			SourceName:      tx.SourceName,
			FormattedDate:   formattedDate,
		})
	}

	return result
}
