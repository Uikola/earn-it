package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// InitClient не закрывает созданное соединение. Надо обязательно его закрыть
func InitClient(ctx context.Context, addr, password string) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	return client, nil
}
