package shop

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

func (r *Repository) ShopItemsByUserID(ctx context.Context, userID int64, isPurchased bool) ([]models.ShopItem, error) {
	q := r.queries(ctx)

	items, err := q.ShopItemsByUserID(ctx, sqlc.ShopItemsByUserIDParams{
		UserID:      userID,
		IsPurchased: isPurchased,
	})
	if err != nil {
		return nil, err
	}

	result := make([]models.ShopItem, 0, len(items))
	for _, item := range items {
		result = append(result, toDomainShopItem(item))
	}

	return result, nil
}

func (r *Repository) ShopItemByID(ctx context.Context, id int64) (models.ShopItem, error) {
	q := r.queries(ctx)

	item, err := q.ShopItemByID(ctx, id)
	if err != nil {
		return models.ShopItem{}, err
	}

	return toDomainShopItem(item), nil
}

func (r *Repository) CreateShopItem(ctx context.Context, userID int64, name string, price int32) (models.ShopItem, error) {
	q := r.queries(ctx)

	item, err := q.CreateShopItem(ctx, sqlc.CreateShopItemParams{
		UserID: userID,
		Name:   name,
		Price:  price,
	})
	if err != nil {
		return models.ShopItem{}, err
	}

	return toDomainShopItem(item), nil
}

func (r *Repository) MarkShopItemPurchased(ctx context.Context, id int64) error {
	q := r.queries(ctx)
	return q.MarkShopItemPurchased(ctx, id)
}

func (r *Repository) CreatePurchase(ctx context.Context, userID, shopItemID int64, pricePaid int32) (models.Purchase, error) {
	q := r.queries(ctx)

	purchase, err := q.CreatePurchase(ctx, sqlc.CreatePurchaseParams{
		UserID:     userID,
		ShopItemID: pgtype.Int8{Int64: shopItemID, Valid: true},
		PricePaid:  pricePaid,
	})
	if err != nil {
		return models.Purchase{}, err
	}

	return toDomainPurchase(purchase), nil
}

func toDomainShopItem(i sqlc.ShopItem) models.ShopItem {
	return models.ShopItem{
		ID:          i.ID,
		UserID:      i.UserID,
		Name:        i.Name,
		Price:       i.Price,
		IsPurchased: i.IsPurchased,
	}
}

func toDomainPurchase(p sqlc.Purchase) models.Purchase {
	return models.Purchase{
		ID:          p.ID,
		UserID:      p.UserID,
		ShopItemID:  p.ShopItemID.Int64,
		PricePaid:   p.PricePaid,
		PurchasedAt: p.PurchasedAt.Time,
	}
}
