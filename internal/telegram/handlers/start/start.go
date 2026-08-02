package start

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"strconv"

	"github.com/nlypage/intele"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/layout"

	"github.com/Uikola/earn-it/internal/models"
	"github.com/Uikola/earn-it/internal/telegram/handlers/helpers"
)

type userRepository interface {
	UserByID(ctx context.Context, id int64) (models.User, error)
	CreateUser(ctx context.Context, id int64, timezone string, rewardWeeklyBonus int32) (models.User, error)
	UpdateUser(ctx context.Context, user models.User) error
}

type Handler struct {
	layout *layout.Layout
	input  *intele.InputManager

	userRepository userRepository
}

func NewHandler(layout *layout.Layout, input *intele.InputManager, userRepository userRepository) *Handler {
	return &Handler{
		layout:         layout,
		input:          input,
		userRepository: userRepository,
	}
}

func (h *Handler) Start(c tele.Context) error {
	ctx := context.Background()

	userID := c.Sender().ID

	_, err := h.userRepository.UserByID(ctx, userID)
	if err == nil {
		return c.Send(
			h.layout.Text(c, "main_menu_text"),
			h.layout.Markup(c, "mainMenu"),
		)
	}

	if !errors.Is(err, sql.ErrNoRows) {
		slog.Error("failed to fetch user", slog.String("err", err.Error()), slog.Int64("userID", userID))
		return c.Send(
			h.layout.Text(c, "technical_issues"),
			h.layout.Markup(c, "mainMenuBack"),
		)
	}

	steps := []helpers.InputStep{
		{
			Name:      "timezone",
			PromptKey: "input_timezone",
			Validator: validateTimezone,
			ErrorKey:  "invalid_timezone",
		},
		{
			Name:      "reward_weekly_bonus",
			PromptKey: "input_reward_weekly_bonus",
			Validator: validateNumber,
			ErrorKey:  "invalid_reward_weekly_bonus",
		},
	}

	results, err := helpers.CollectInput(c, h.input, h.layout, steps, nil, true)
	if err != nil {
		if errors.Is(err, helpers.ErrCanceled) {
			return nil
		}
		slog.Error("failed to collect input", slog.String("err", err.Error()))
		return c.Send(
			h.layout.Text(c, "technical_issues"),
			h.layout.Markup(c, "mainMenuBack"),
		)
	}

	timezone := results["timezone"]
	rewardWeeklyBonus, _ := strconv.ParseInt(results["reward_weekly_bonus"], 10, 32)

	_, err = h.userRepository.CreateUser(ctx, userID, timezone, int32(rewardWeeklyBonus))
	if err != nil {
		slog.Error("failed to create user", slog.String("err", err.Error()), slog.Int64("userID", userID))
		return c.Send(
			h.layout.Text(c, "technical_issues"),
			h.layout.Markup(c, "mainMenuBack"),
		)
	}

	return c.Send(
		h.layout.Text(c, "main_menu_text"),
		h.layout.Markup(c, "mainMenu"),
	)
}

func (h *Handler) MainMenu(c tele.Context) error {
	ctx := context.Background()

	_, err := h.userRepository.UserByID(ctx, c.Sender().ID)
	if err == nil {
		return c.Edit(
			h.layout.Text(c, "main_menu_text"),
			h.layout.Markup(c, "mainMenu"),
		)
	}

	if !errors.Is(err, sql.ErrNoRows) {
		slog.Error("failed to fetch user", slog.String("err", err.Error()), slog.Int64("userID", c.Sender().ID))
		return c.Edit(
			h.layout.Text(c, "technical_issues"),
			h.layout.Markup(c, "mainMenuBack"),
		)
	}

	return c.Edit(
		h.layout.Text(c, "unauthorized"),
		h.layout.Markup(c, "core:hide"),
	)
}
