package transaction

import (
	"context"

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

func (r *Repository) CreateTransaction(ctx context.Context, userID int64, amount int32, source, sourceName string) (models.Transaction, error) {
	q := r.queries(ctx)

	tx, err := q.CreateTransaction(ctx, sqlc.CreateTransactionParams{
		UserID:     userID,
		Amount:     amount,
		Source:     source,
		SourceName: sourceName,
	})
	if err != nil {
		return models.Transaction{}, err
	}

	return toDomainTransaction(tx), nil
}

func (r *Repository) TotalIncome(ctx context.Context, userID int64) (int32, error) {
	q := r.queries(ctx)

	result, err := q.TotalIncome(ctx, userID)
	if err != nil {
		return 0, err
	}

	if result == nil {
		return 0, nil
	}

	switch v := result.(type) {
	case int64:
		return int32(v), nil
	case int32:
		return v, nil
	case float64:
		return int32(v), nil
	default:
		return 0, nil
	}
}

func (r *Repository) TotalExpense(ctx context.Context, userID int64) (int32, error) {
	q := r.queries(ctx)

	result, err := q.TotalExpense(ctx, userID)
	if err != nil {
		return 0, err
	}

	if result == nil {
		return 0, nil
	}

	switch v := result.(type) {
	case int64:
		return int32(v), nil
	case int32:
		return v, nil
	case float64:
		return int32(v), nil
	default:
		return 0, nil
	}
}

func (r *Repository) RecentTransactionsWithDetails(ctx context.Context, userID int64, limit int32) ([]models.TransactionWithDetails, error) {
	q := r.queries(ctx)

	rows, err := q.RecentTransactionsWithDetails(ctx, sqlc.RecentTransactionsWithDetailsParams{
		UserID: userID,
		Limit:  limit,
	})
	if err != nil {
		return nil, err
	}

	result := make([]models.TransactionWithDetails, 0, len(rows))
	for _, row := range rows {
		result = append(result, models.TransactionWithDetails{
			ID:         row.ID,
			Amount:     row.Amount,
			Source:     row.Source,
			SourceName: row.SourceName,
			CreatedAt:  row.CreatedAt.Time,
		})
	}

	return result, nil
}

func toDomainTransaction(t sqlc.Transaction) models.Transaction {
	return models.Transaction{
		ID:         t.ID,
		UserID:     t.UserID,
		Amount:     t.Amount,
		Source:     t.Source,
		SourceName: t.SourceName,
		CreatedAt:  t.CreatedAt.Time,
	}
}
