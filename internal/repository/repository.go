package repository

import (
	"context"

	dice "github.com/dicedb/go-dice"
)

type Repository interface {
	Get(ctx context.Context, req *GetRequest) (string, error)
	Set(ctx context.Context, req *SetRequest) error
	Delete(ctx context.Context, req *DeleteRequest) error
}

type repository struct {
	client *dice.Client
}

func NewRepository(client *dice.Client) Repository {
	return &repository{client: client}
}

func (r *repository) Get(ctx context.Context, req *GetRequest) (string, error) {
	return r.client.Get(ctx, req.Key).Result()
}

func (r *repository) Set(ctx context.Context, req *SetRequest) error {
	return r.client.Set(ctx, req.Key, req.Value, 0).Err()
}

func (r *repository) Delete(ctx context.Context, req *DeleteRequest) error {
	return r.client.Del(ctx, req.Keys...).Err()
}
