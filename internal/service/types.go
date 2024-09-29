package service

import (
	"context"
	"server/internal/repository"
)

// Service is service layer client interface that should have all the methods
// exposed to handlers
type Service interface {
	Get(ctx context.Context, req *GetRequest) (*GetResponse, error)
	Set(ctx context.Context, req *SetRequest) (*SetResponse, error)
	Delete(ctx context.Context, req *DeleteRequest) (*DeleteResponse, error)
}

// service is private struct and concrete implementation that implements the
// Service client level interface
type service struct {
	repo repository.Repository
}

// NewService instantiates the service layer it takes Repository layer's client
// interface as dependency
func NewService(repo repository.Repository) Service {
	return &service{repo: repo}
}

// Service layer response and request models for GET

type GetRequest struct {
	Key string
}

type GetResponse struct {
	Value string
}

// Service layer response and request models for SET

type SetRequest struct {
	Key   string
	Value string
}

type SetResponse struct {
	Success string
}

// Service layer response and request models for DELETE

type DeleteRequest struct {
	Keys []string
}

type DeleteResponse struct {
	Success string
}
