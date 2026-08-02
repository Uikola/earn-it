package transaction

import (
	"context"

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

func (r *Repository) queries(ctx context.Context) *sqlc.Queries {
	executor := postgres.GetQueryExecutor(ctx, r.pool)
	return sqlc.New(executor)
}

func (r *Repository) CreateTransaction(ctx context.Context, userID int64, amount int32, source string, sourceID int64) (models.Transaction, error) {
	q := r.queries(ctx)

	tx, err := q.CreateTransaction(ctx, sqlc.CreateTransactionParams{
		UserID:   userID,
		Amount:   amount,
		Source:   source,
		SourceID: pgtype.Int8{Int64: sourceID, Valid: sourceID > 0},
	})
	if err != nil {
		return models.Transaction{}, err
	}

	return toDomainTransaction(tx), nil
}

func toDomainTransaction(t sqlc.Transaction) models.Transaction {
	return models.Transaction{
		ID:        t.ID,
		UserID:    t.UserID,
		Amount:    t.Amount,
		Source:    t.Source,
		SourceID:  t.SourceID.Int64,
		CreatedAt: t.CreatedAt.Time,
	}
}
