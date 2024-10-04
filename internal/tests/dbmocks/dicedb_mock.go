package db

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type DiceDBMock struct {
	data  map[string]string
	mutex sync.Mutex
}

func NewDiceDBMock() *DiceDBMock {
	return &DiceDBMock{
		data: make(map[string]string),
	}
}

func (db *DiceDBMock) Get(ctx context.Context, key string) (string, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	val, exists := db.data[key]
	if !exists {
		return "", nil
	}
	return val, nil
}

func (db *DiceDBMock) Set(ctx context.Context, key, value string, expiration time.Duration) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	db.data[key] = value
	return nil
}

func (db *DiceDBMock) Incr(ctx context.Context, key string) (int64, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	val, exists := db.data[key]
	var count int64
	if exists {
		if _, err := fmt.Sscanf(val, "%d", &count); err != nil {
			return 0, fmt.Errorf("error parsing value for key %s: %w", key, err)
		}
	}

	count++
	db.data[key] = fmt.Sprintf("%d", count)
	return count, nil
}

func (db *DiceDBMock) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return nil
}
