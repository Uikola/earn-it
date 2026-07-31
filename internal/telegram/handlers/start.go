package handlers

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Uikola/earn-it/internal/models"
	"github.com/google/martian/log"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/layout"
)

type userRepository interface {
	UserByID(ctx context.Context, id int64) (models.User, error)
	CreateUser(ctx context.Context, id int64, timezone string, rewardWeeklyBonus int32) (models.User, error)
	UpdateUser(ctx context.Context, user models.User) error
}

type StartHandler struct {
	layout *layout.Layout

	userRepository userRepository
}

func NewStartHandler(layout *layout.Layout, userRepository userRepository) *StartHandler {
	return &StartHandler{layout: layout, userRepository: userRepository}
}

func (h *StartHandler) Start(c tele.Context) error {
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
		log.Errorf("failed to fetch user: %s", err)
		return c.Send(
			h.layout.Text(c, "technical_issues"),
			h.layout.Markup(c, "mainMenuBack"),
		)
	}

	// TODO: Сделать этап регистрации, где буду предлогать пользователю захардкоженные данные
	_, err = h.userRepository.CreateUser(ctx, userID, "Europe/Moscow", 40)
	if err != nil {
		log.Errorf("failed to create user: %s", err)
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

func (h *StartHandler) MainMenu(c tele.Context) error {
	ctx := context.Background()

	_, err := h.userRepository.UserByID(ctx, c.Sender().ID)
	if err == nil {
		return c.Edit(
			h.layout.Text(c, "main_menu_text"),
			h.layout.Markup(c, "mainMenu"),
		)
	}

	if !errors.Is(err, sql.ErrNoRows) {
		log.Errorf("failed to fetch user: %s", err)
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

func (h *StartHandler) Clear(c tele.Context) error {
	// TODO: Надо как-то оптимально хранить idшники всех сообщений пользовател, получать их и проходясь по списку удалять
	// Пока смысла нет, сделаю позже нейронкой
	return nil
}
