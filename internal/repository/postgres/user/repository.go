package user

import (
	"context"

	"github.com/Uikola/earn-it/internal/models"
	"github.com/Uikola/earn-it/internal/repository/postgres"
	"github.com/Uikola/earn-it/internal/repository/postgres/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
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

func (r *Repository) UserByID(ctx context.Context, id int64) (models.User, error) {
	q := r.queries(ctx)

	user, err := q.UserByID(ctx, id)
	if err != nil {
		return models.User{}, err
	}

	return toDomainUser(user), nil
}

func (r *Repository) CreateUser(ctx context.Context, id int64, timezone string, rewardWeeklyBonus int32) (models.User, error) {
	q := r.queries(ctx)

	row, err := q.CreateUser(ctx, sqlc.CreateUserParams{
		ID:                id,
		Timezone:          pgtype.Text{String: timezone, Valid: true},
		RewardWeeklyBonus: rewardWeeklyBonus,
	})
	if err != nil {
		return models.User{}, err
	}

	return models.User{
		ID:                row.ID,
		Timezone:          row.Timezone.String,
		Balance:           row.Balance.Int32,
		RewardWeeklyBonus: row.RewardWeeklyBonus,
		CreatedAt:         row.CreatedAt.Time,
	}, nil
}

func (r *Repository) UpdateUser(ctx context.Context, user models.User) error {
	q := r.queries(ctx)

	return q.UpdateUser(ctx, sqlc.UpdateUserParams{
		ID:                user.ID,
		Timezone:          pgtype.Text{String: user.Timezone, Valid: true},
		Balance:           pgtype.Int4{Int32: user.Balance, Valid: true},
		RewardWeeklyBonus: user.RewardWeeklyBonus,
	})
}

func toDomainUser(u sqlc.User) models.User {
	return models.User{
		ID:                u.ID,
		Timezone:          u.Timezone.String,
		Balance:           u.Balance.Int32,
		RewardWeeklyBonus: u.RewardWeeklyBonus,
		CreatedAt:         u.CreatedAt.Time,
	}
}
