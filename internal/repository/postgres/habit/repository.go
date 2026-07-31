package habit

import (
	"context"

	"github.com/Uikola/earn-it/internal/models"
	"github.com/Uikola/earn-it/internal/repository/postgres/sqlc"
	"github.com/jackc/pgx/v5"
)

type Repository struct {
	q *sqlc.Queries
}

func NewRepository(db sqlc.DBTX) *Repository {
	return &Repository{q: sqlc.New(db)}
}

func WithTx(tx pgx.Tx) *Repository {
	return &Repository{q: sqlc.New(tx)}
}

func (r *Repository) HabitsByUserID(ctx context.Context, userID int64) ([]models.Habit, error) {
	habits, err := r.q.HabitsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	domainHabits := make([]models.Habit, 0, len(habits))
	for _, habit := range habits {
		domainHabits = append(domainHabits, toDomainHabit(habit))
	}

	return domainHabits, nil
}

func (r *Repository) CreateHabit(ctx context.Context, userID int64, name string, weaklyGoal int32, rewardPerExecute int32) (models.Habit, error) {
	habit, err := r.q.CreateHabit(ctx, sqlc.CreateHabitParams{
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
