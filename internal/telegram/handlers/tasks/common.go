package tasks

import (
	"context"
	"log/slog"
	"time"

	"github.com/nlypage/intele"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/layout"

	"github.com/Uikola/earn-it/internal/models"
	"github.com/Uikola/earn-it/internal/repository/postgres"
)

type taskRepository interface {
	CreateTask(ctx context.Context, userID int64, title string, scheduledDate time.Time, rewardValue int32) (models.Task, error)
	TaskByID(ctx context.Context, taskID int64) (models.Task, error)
	TasksByUserAndDateRange(ctx context.Context, userID int64, from, to time.Time) ([]models.Task, error)
	TasksByUserID(ctx context.Context, userID int64) ([]models.Task, error)
	UpdateTask(ctx context.Context, taskID int64, task models.Task) error
	RescheduleExpiredTasks(ctx context.Context, userID int64, today time.Time) error
	DeleteTask(ctx context.Context, taskID int64) error
}

type userRepository interface {
	UserByID(ctx context.Context, id int64) (models.User, error)
	UpdateUser(ctx context.Context, user models.User) error
}

type Handler struct {
	layout         *layout.Layout
	input          *intele.InputManager
	transactor     postgres.Transactor
	taskRepository taskRepository
	userRepository userRepository
}

func NewHandler(
	layout *layout.Layout,
	input *intele.InputManager,
	transactor postgres.Transactor,
	taskRepository taskRepository,
	userRepository userRepository,
) *Handler {
	return &Handler{
		layout:         layout,
		input:          input,
		transactor:     transactor,
		taskRepository: taskRepository,
		userRepository: userRepository,
	}
}

type taskToPrint struct {
	ID            int64
	Title         string
	ScheduledDate string
	RewardValue   int32
}

func (h *Handler) userByIDWithProcessedError(ctx context.Context, c tele.Context, userID int64) *models.User {
	user, err := h.userRepository.UserByID(ctx, userID)
	if err != nil {
		slog.Error("failed UserByID", slog.String("err", err.Error()), slog.Int64("userID", userID))
		if err := c.Edit(
			h.layout.Text(c, "technical_issues"),
			h.layout.Markup(c, "mainMenuBack"),
		); err != nil {
			return nil
		}
		return nil
	}
	return &user
}

func (h *Handler) tasksToPrint(timezone string, tasks []models.Task) []taskToPrint {
	tasksToPrint := make([]taskToPrint, 0, len(tasks))

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}

	now := time.Now().In(loc)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	tomorrow := today.AddDate(0, 0, 1)

	for _, task := range tasks {
		taskDate := task.ScheduledDate.In(loc)
		taskDateOnly := time.Date(taskDate.Year(), taskDate.Month(), taskDate.Day(), 0, 0, 0, 0, loc)

		var dateStr string
		switch {
		case taskDateOnly.Equal(today):
			dateStr = "Сегодня"
		case taskDateOnly.Equal(tomorrow):
			dateStr = "Завтра"
		default:
			dateStr = taskDateOnly.Format("02.01.2006")
		}

		tasksToPrint = append(tasksToPrint, taskToPrint{
			ID:            task.ID,
			Title:         task.Title,
			ScheduledDate: dateStr,
			RewardValue:   task.RewardValue,
		})
	}

	return tasksToPrint
}
