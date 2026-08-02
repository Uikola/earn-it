package habits

import (
	"context"
	"time"

	"github.com/google/martian/log"
	"github.com/nlypage/intele"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/layout"

	"github.com/Uikola/earn-it/internal/models"
	"github.com/Uikola/earn-it/internal/repository/postgres"
)

type habitRepository interface {
	HabitsByUserID(ctx context.Context, userID int64) ([]models.Habit, error)
	HabitByID(ctx context.Context, habitID int64) (models.Habit, error)
	CreateHabit(ctx context.Context, userID int64, name string, weeklyGoal, rewardPerExecute int32) (models.Habit, error)
	DeleteHabit(ctx context.Context, habitID int64) error

	CreateHabitLog(ctx context.Context, habitID int64) error
	HabitLogsForWeek(ctx context.Context, habitID int64, weekStart time.Time) ([]models.HabitLog, error)
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
	WeeklyGoalDone      int32
	WeeklyGoal          int32
	WeeklyGoalCompleted bool
	RewardPerExecute    int32
	ProgressBar         string
}

func progressBar(done, goal int32) string {
	if goal <= 0 {
		return ""
	}
	filled := done
	if filled > goal {
		filled = goal
	}
	runes := make([]rune, 0, goal)
	for i := int32(0); i < filled; i++ {
		runes = append(runes, '🟩')
	}
	for i := filled; i < goal; i++ {
		runes = append(runes, '⬜')
	}
	return string(runes)
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

func (h *Handler) habitsToPrint(ctx context.Context, c tele.Context, weekStart time.Time, habits []models.Habit) []habitToPrint {
	habitsToPrint := make([]habitToPrint, 0, len(habits))
	for _, habit := range habits {
		habitLogsForWeek, err := h.habitsRepository.HabitLogsForWeek(ctx, habit.ID, weekStart)
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
			WeeklyGoalDone:      int32(len(habitLogsForWeek)),
			WeeklyGoal:          habit.WeeklyGoal,
			WeeklyGoalCompleted: int32(len(habitLogsForWeek)) >= habit.WeeklyGoal,
			RewardPerExecute:    habit.RewardPerExecute,
			ProgressBar:         progressBar(int32(len(habitLogsForWeek)), habit.WeeklyGoal),
		})
	}

	return habitsToPrint
}
