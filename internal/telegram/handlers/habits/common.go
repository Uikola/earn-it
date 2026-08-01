package habits

import (
	"context"

	"github.com/Uikola/earn-it/internal/models"
	"github.com/Uikola/earn-it/internal/repository/postgres"
	"github.com/Uikola/earn-it/internal/repository/postgres/sqlc"
	"github.com/google/martian/log"
	"github.com/nlypage/intele"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/layout"
)

type habitRepository interface {
	HabitsByUserID(ctx context.Context, userID int64) ([]models.Habit, error)
	HabitByID(ctx context.Context, habitID int64) (models.Habit, error)
	CreateHabit(ctx context.Context, userID int64, name string, weaklyGoal, rewardPerExecute int32) (models.Habit, error)
	DeleteHabit(ctx context.Context, habitID int64) error

	CreateHabitLog(ctx context.Context, habitID int64) error
	HabitLogsForWeek(ctx context.Context, habitID int64, timezoneStr string) ([]sqlc.HabitLog, error)
}

type userRepository interface {
	UserByID(ctx context.Context, id int64) (models.User, error)
	UpdateUser(ctx context.Context, user models.User) error
}

type Handler struct {
	layout *layout.Layout
	input  *intele.InputManager

	transactor       postgres.Transactor
	habitsRepository habitRepository
	userRepository   userRepository
}

func NewHandler(
	layout *layout.Layout,
	input *intele.InputManager,
	transactor postgres.Transactor,
	habitsRepository habitRepository,
	userRepository userRepository,
) *Handler {
	return &Handler{
		layout: layout,
		input:  input,

		transactor:       transactor,
		habitsRepository: habitsRepository,
		userRepository:   userRepository,
	}
}

type habitToPrint struct {
	Name                string
	WeaklyGoalDone      int32
	WeaklyGoal          int32
	WeaklyGoalCompleted bool
	RewardPerExecute    int32
}

func (h *Handler) userByIDWithProcessedError(ctx context.Context, c tele.Context, userID int64, action string) *models.User {
	user, err := h.userRepository.UserByID(ctx, userID)
	if err != nil {
		log.Errorf("failed to fetch user: %v", err)

		switch action {
		case "send":
			if err := c.Send(
				h.layout.Text(c, "technical_issues"),
				h.layout.Markup(c, "mainMenuBack"),
			); err != nil {
				log.Errorf("SendError while userByIDWithProcessedError")
				return nil
			}
		case "edit":
			if err := c.Edit(
				h.layout.Text(c, "technical_issues"),
				h.layout.Markup(c, "mainMenuBack"),
			); err != nil {
				log.Errorf("EditError while userByIDWithProcessedError")
				return nil
			}
		}

		return nil
	}

	return &user
}

func (h *Handler) habitsByUserIDWithProcessedError(ctx context.Context, c tele.Context, userID int64, action string) []models.Habit {
	habits, err := h.habitsRepository.HabitsByUserID(ctx, userID)
	if err != nil {
		log.Errorf("failed to fetch habits: %v", err)

		switch action {
		case "send":
			if err := c.Send(
				h.layout.Text(c, "technical_issues"),
				h.layout.Markup(c, "mainMenuBack"),
			); err != nil {
				log.Errorf("SendError while habitsByUserIDWithProcessedError")
				return nil
			}
		case "edit":
			if err := c.Edit(
				h.layout.Text(c, "technical_issues"),
				h.layout.Markup(c, "mainMenuBack"),
			); err != nil {
				log.Errorf("EditError while habitsByUserIDWithProcessedError")
				return nil
			}
		}
	}

	return habits
}

func (h *Handler) habitsToPrint(ctx context.Context, c tele.Context, timezone string, habits []models.Habit) []habitToPrint {
	habitsToPrint := make([]habitToPrint, 0, len(habits))
	for _, habit := range habits {
		habitLogsForWeek, err := h.habitsRepository.HabitLogsForWeek(ctx, habit.ID, timezone)
		if err != nil {
			log.Errorf("failed to fetch habit logs for week: %v", err)
			if err := c.Edit(
				h.layout.Text(c, "technical_issues"),
				h.layout.Markup(c, "mainMenuBack"),
			); err != nil {
				log.Errorf("EditError while habitsToPrint")
				return nil
			}
		}

		habitsToPrint = append(habitsToPrint, habitToPrint{
			Name:                habit.Name,
			WeaklyGoalDone:      int32(len(habitLogsForWeek)),
			WeaklyGoal:          habit.WeaklyGoal,
			WeaklyGoalCompleted: int32(len(habitLogsForWeek)) >= habit.WeaklyGoal,
			RewardPerExecute:    habit.RewardPerExecute,
		})
	}

	return habitsToPrint
}
