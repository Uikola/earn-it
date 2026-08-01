package scheduler

import (
	"context"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/Uikola/earn-it/internal/models"
	"github.com/Uikola/earn-it/internal/repository/postgres"
)

type userRepository interface {
	Users(ctx context.Context) ([]models.User, error)
	UserByID(ctx context.Context, id int64) (models.User, error)
	UpdateUser(ctx context.Context, user models.User) error
}

type habitRepository interface {
	HabitsByUserID(ctx context.Context, userID int64) ([]models.Habit, error)
	HabitLogsForWeek(ctx context.Context, habitID int64, weekStart time.Time) ([]models.HabitLog, error)
	WeeklyBonus(ctx context.Context, userID int64, weekStart time.Time) (models.WeeklyBonus, error)
	CreateWeeklyBonus(ctx context.Context, userID int64, weekStart time.Time) error
}

type Scheduler struct {
	cron            *cron.Cron
	transactor      postgres.Transactor
	userRepository  userRepository
	habitRepository habitRepository
}

func New(
	transactor postgres.Transactor,
	userRepository userRepository,
	habitRepository habitRepository,
) *Scheduler {
	return &Scheduler{
		cron:            cron.New(),
		transactor:      transactor,
		userRepository:  userRepository,
		habitRepository: habitRepository,
	}
}

func (s *Scheduler) Start() {
	_, err := s.cron.AddFunc("@every 1h", s.processWeeklyBonuses)
	if err != nil {
		slog.Error("failed to add weekly bonus job", slog.String("error", err.Error()))
		return
	}

	s.cron.Start()
	slog.Info("scheduler started")
}

func (s *Scheduler) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
	slog.Info("scheduler stopped")
}
