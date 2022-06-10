package store

import (
	"context"
	"dcard-pretest/pkg/model"

	"github.com/go-redis/redis/v8"
)

func NewRedis() *Store {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	store := Store{client: redisClient}

	return &store
}

type Store struct {
	client *redis.Client
}

func (s *Store) Insert(ctx context.Context, key string, score model.Score) error {
	err := s.client.ZAdd(ctx, key, &redis.Z{
		Score:  score.Score,
		Member: score.ClientId,
	}).Err()

	return err
}

func (s *Store) Top10(ctx context.Context, key string) ([]*model.Score, error) {
	scores, err := s.client.ZRevRangeWithScores(ctx, key, 0, 9).Result()
	if err != nil {
		return make([]*model.Score, 0), err
	}

	result := make([]*model.Score, len(scores))

	for i, v := range scores {
		result[i] = &model.Score{
			ClientId: v.Member.(string),
			Score:    v.Score,
		}
	}
	return result, nil
}

func (s *Store) Reset(ctx context.Context, key string) error {
	err := s.client.Del(ctx, key).Err()
	return err
}

func (s *Store) Close() error {
	err := s.client.Close()
	return err
}
