package habit

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Uikola/earn-it/internal/models"
	"github.com/Uikola/earn-it/internal/repository/postgres"
	"github.com/Uikola/earn-it/internal/repository/postgres/sqlc"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// вспомогательный метод для получения sqlc.Queries с executor из контекста
func (r *Repository) queries(ctx context.Context) *sqlc.Queries {
	executor := postgres.GetQueryExecutor(ctx, r.pool)
	return sqlc.New(executor) // executor реализует интерфейс sqlc.DBTX
}

func (r *Repository) HabitsByUserID(ctx context.Context, userID int64) ([]models.Habit, error) {
	q := r.queries(ctx)

	habits, err := q.HabitsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	domainHabits := make([]models.Habit, 0, len(habits))
	for _, habit := range habits {
		domainHabits = append(domainHabits, toDomainHabit(habit))
	}

	return domainHabits, nil
}

func (r *Repository) HabitByID(ctx context.Context, habitID int64) (models.Habit, error) {
	q := r.queries(ctx)

	habit, err := q.HabitByID(ctx, habitID)
	if err != nil {
		return models.Habit{}, err
	}

	return toDomainHabit(habit), nil
}

func (r *Repository) CreateHabit(ctx context.Context, userID int64, name string, weaklyGoal, rewardPerExecute int32) (models.Habit, error) {
	q := r.queries(ctx)

	habit, err := q.CreateHabit(ctx, sqlc.CreateHabitParams{
		UserID:           userID,
		Name:             name,
		WeeklyGoal:       weaklyGoal,
		RewardPerExecute: rewardPerExecute,
	})
	if err != nil {
		return models.Habit{}, err
	}

	return toDomainHabit(habit), nil
}

func (r *Repository) DeleteHabit(ctx context.Context, habitID int64) error {
	q := r.queries(ctx)

	return q.DeleteHabit(ctx, habitID)
}

func (r *Repository) CreateHabitLog(ctx context.Context, habitID int64) error {
	q := r.queries(ctx)

	_, err := q.CreateHabitLog(ctx, habitID)
	return err
}

func (r *Repository) HabitLogsForWeek(
	ctx context.Context,
	habitID int64,
	timezoneStr string,
) ([]sqlc.HabitLog, error) {
	q := r.queries(ctx)

	loc, err := time.LoadLocation(timezoneStr)
	if err != nil {
		return nil, fmt.Errorf("invalid timezone %q: %w", timezoneStr, err)
	}

	now := time.Now().In(loc)

	weekday := now.Weekday()
	offset := (int(weekday) - int(time.Monday) + 7) % 7
	weekStartLocal := now.AddDate(0, 0, -offset).Truncate(24 * time.Hour)

	weekStartUTC := weekStartLocal.UTC()

	weekStartParam := pgtype.Timestamptz{
		Time:  weekStartUTC,
		Valid: true,
	}

	logs, err := q.HabitLogsForWeek(ctx, sqlc.HabitLogsForWeekParams{
		HabitID:    habitID,
		ExecutedAt: weekStartParam,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get habit logs: %w", err)
	}

	return logs, nil
}

func toDomainHabit(h sqlc.Habit) models.Habit {
	return models.Habit{
		ID:               h.ID,
		UserID:           h.UserID,
		Name:             h.Name,
		WeaklyGoal:       h.WeeklyGoal,
		RewardPerExecute: h.RewardPerExecute,
		CreatedAt:        h.CreatedAt.Time,
	}
}
