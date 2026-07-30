package user

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Uikola/earn-it/internal/models"
	"github.com/Uikola/earn-it/internal/repository/postgres/sqlc"
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

func (r *Repository) UserByID(ctx context.Context, id int) (models.User, error) {
	user, err := r.q.UserByID(ctx, int64(id))
	if err != nil {
		return models.User{}, err
	}
	return toDomainUser(user), nil
}

func (r *Repository) CreateUser(ctx context.Context, id int64, timezone string, rewardWeeklyBonus int32) (models.User, error) {
	row, err := r.q.CreateUser(ctx, sqlc.CreateUserParams{
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
	return r.q.UpdateUser(ctx, sqlc.UpdateUserParams{
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
