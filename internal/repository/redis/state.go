package redis

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type StateRepository struct {
	client *redis.Client
}

func NewStateRepository(client *redis.Client) *StateRepository {
	return &StateRepository{client: client}
}

func (s *StateRepository) Set(userID int64, state string, expiration time.Duration) error {
	return s.client.Set(context.Background(), strconv.FormatInt(userID, 10), state, expiration).Err()
}

func (s *StateRepository) Get(userID int64) (string, error) {
	return s.client.Get(context.Background(), strconv.FormatInt(userID, 10)).Result()
}

func (s *StateRepository) Delete(userID int64) {
	s.client.Del(context.Background(), strconv.FormatInt(userID, 10))
}
