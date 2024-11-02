package storage

import (
	"captcha-bot/internal/app/logic"
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type UserRedisRepo struct {
	client *redis.Client
	ttl    time.Duration
}

func NewUserRedisRepo(stateTTL time.Duration, redisClient *redis.Client) *UserRedisRepo {
	return &UserRedisRepo{
		client: redisClient,
		ttl:    stateTTL,
	}
}

func (r *UserRedisRepo) GetUserData(ctx context.Context, userID int64, chatID int64) (*logic.UserData, error) {
	key := strconv.FormatInt(userID, 10) + strconv.FormatInt(chatID, 10)
	result, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, logic.ErrStateNotFound
	} else if err != nil {
		return nil, err
	}

	userData, err := logic.UnmarshalUserData([]byte(result))
	if err != nil {
		return nil, errors.New("state decode error")
	}

	if userData.Expired() {
		r.client.Del(ctx, key)
		return nil, logic.ErrStateNotFound
	}

	return userData, nil
}

func (r *UserRedisRepo) Put(ctx context.Context, userData *logic.UserData) error {
	if userData.UserID == 0 {
		return errors.New("couldn't put userData with empty UserID")
	}
	if userData.ChatID == 0 {
		return errors.New("couldn't put userData with empty ChatID")
	}

	// expiration := time.Duration(0)
	// if r.ttl > 0 {
	// 	expiration = r.ttl * time.Second
	// }
	// userData.Expiration = time.Now().Add(expiration).UnixNano()

	data, err := logic.MarshalUserData(userData)
	if err != nil {
		return err
	}

	key := strconv.FormatInt(userData.UserID, 10) + strconv.FormatInt(userData.ChatID, 10)
	err = r.client.Set(ctx, key, data, redis.KeepTTL).Err()
	if err != nil {
		return err
	}

	return nil
}

func (r *UserRedisRepo) Remove(ctx context.Context, userID int64, chatID int64) {
	key := strconv.FormatInt(userID, 10) + strconv.FormatInt(chatID, 10)
	r.client.Del(ctx, key)
}
