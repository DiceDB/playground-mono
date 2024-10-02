package db

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type InMemoryDiceDB struct {
	data  map[string]string
	mutex sync.Mutex
}

func NewInMemoryDiceDB() *InMemoryDiceDB {
	return &InMemoryDiceDB{
		data: make(map[string]string),
	}
}

func (db *InMemoryDiceDB) Get(ctx context.Context, key string) (string, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	val, exists := db.data[key]
	if !exists {
		return "", nil
	}
	return val, nil
}

func (db *InMemoryDiceDB) Set(ctx context.Context, key, value string, expiration time.Duration) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	db.data[key] = value
	return nil
}

func (db *InMemoryDiceDB) Incr(ctx context.Context, key string) (int64, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	val, exists := db.data[key]
	var count int64
	if exists {
		fmt.Sscanf(val, "%d", &count)
	}

	count++
	db.data[key] = fmt.Sprintf("%d", count)
	return count, nil
}

func (db *InMemoryDiceDB) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return nil
}
