package service

import (
	"context"
	"errors"
	"server/internal/repository"
)

func (s *service) Get(ctx context.Context, req *GetRequest) (*GetResponse, error) {
	if req.Key == "" {
		return &GetResponse{}, errors.New("key is required")
	}

	value, err := s.repo.Get(ctx, &repository.GetRequest{
		Key: req.Key,
	})
	if err != nil {
		return &GetResponse{}, err
	}

	return &GetResponse{Value: value}, nil
}

func (s *service) Set(ctx context.Context, req *SetRequest) (*SetResponse, error) {
	if req.Key == "" || req.Value == "" {
		return nil, errors.New("key and value are required")
	}

	err := s.repo.Set(ctx, &repository.SetRequest{
		Key:   req.Key,
		Value: req.Value,
	})
	if err != nil {
		return nil, err
	}

	return &SetResponse{Success: "OK"}, nil
}

func (s *service) Delete(ctx context.Context, req *DeleteRequest) (*DeleteResponse, error) {
	if len(req.Keys) == 0 {
		return nil, errors.New("at least one key is required")
	}

	err := s.repo.Delete(ctx, &repository.DeleteRequest{
		Keys: req.Keys,
	})
	if err != nil {
		return nil, err
	}

	return &DeleteResponse{Success: "OK"}, nil
}
