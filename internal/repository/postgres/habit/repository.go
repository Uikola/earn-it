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

func (r *Repository) CreateHabit(ctx context.Context, userID int64, name string, weeklyGoal, rewardPerExecute int32) (models.Habit, error) {
	q := r.queries(ctx)

	habit, err := q.CreateHabit(ctx, sqlc.CreateHabitParams{
		UserID:           userID,
		Name:             name,
		WeeklyGoal:       weeklyGoal,
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
	weekStart time.Time,
) ([]models.HabitLog, error) {
	q := r.queries(ctx)

	weekStartParam := pgtype.Timestamptz{
		Time:  weekStart,
		Valid: true,
	}

	logs, err := q.HabitLogsForWeek(ctx, sqlc.HabitLogsForWeekParams{
		HabitID:    habitID,
		ExecutedAt: weekStartParam,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get habit logs: %w", err)
	}

	domainLogs := make([]models.HabitLog, 0, len(logs))
	for _, log := range logs {
		domainLogs = append(domainLogs, toDomainHabitLog(log))
	}

	return domainLogs, nil
}

func (r *Repository) WeeklyBonus(ctx context.Context, userID int64, weekStart time.Time) (models.WeeklyBonus, error) {
	q := r.queries(ctx)

	weekStartParam := pgtype.Date{
		Time:  weekStart,
		Valid: true,
	}

	bonus, err := q.WeeklyBonus(ctx, sqlc.WeeklyBonusParams{
		UserID:    userID,
		WeekStart: weekStartParam,
	})
	if err != nil {
		return models.WeeklyBonus{}, err
	}

	return toDomainWeeklyBonus(bonus), nil
}

func (r *Repository) CreateWeeklyBonus(ctx context.Context, userID int64, weekStart time.Time) error {
	q := r.queries(ctx)

	weekStartParam := pgtype.Date{
		Time:  weekStart,
		Valid: true,
	}

	return q.CreateWeeklyBonus(ctx, sqlc.CreateWeeklyBonusParams{
		UserID:    userID,
		WeekStart: weekStartParam,
	})
}

func toDomainHabit(h sqlc.Habit) models.Habit {
	return models.Habit{
		ID:               h.ID,
		UserID:           h.UserID,
		Name:             h.Name,
		WeeklyGoal:       h.WeeklyGoal,
		RewardPerExecute: h.RewardPerExecute,
		CreatedAt:        h.CreatedAt.Time,
	}
}

func toDomainHabitLog(l sqlc.HabitLog) models.HabitLog {
	return models.HabitLog{
		ID:         l.ID,
		HabitID:    l.HabitID,
		ExecutedAt: l.ExecutedAt.Time,
	}
}

func toDomainWeeklyBonus(b sqlc.WeeklyBonuse) models.WeeklyBonus {
	return models.WeeklyBonus{
		ID:        b.ID,
		UserID:    b.UserID,
		WeekStart: b.WeekStart.Time,
		ClaimedAt: b.ClaimedAt.Time,
	}
}
