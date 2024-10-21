package server

import (
	"context"
	"errors"
	"log/slog"
	"server/internal/db"
	"server/internal/server/utils"
	"strconv"
	"sync"
	"time"

	"github.com/dicedb/dicedb-go"
)

type CleanupManager struct {
	diceDBAdminClient *db.DiceDB
	diceDBClient      *db.DiceDB
	cronFrequency     time.Duration
}

func NewCleanupManager(diceDBAdminClient *db.DiceDB,
	diceDBClient *db.DiceDB, cronFrequency time.Duration) *CleanupManager {
	return &CleanupManager{
		diceDBAdminClient: diceDBAdminClient,
		diceDBClient:      diceDBClient,
		cronFrequency:     cronFrequency,
	}
}

func (c *CleanupManager) Run(ctx context.Context) {
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		c.start(ctx)
	}()
}

func (c *CleanupManager) start(ctx context.Context) {
	ticker := time.NewTicker(c.cronFrequency)
	defer ticker.Stop()

	// Get the last cron run time
	resp := c.diceDBAdminClient.Client.Get(ctx, utils.LastCronCleanupTimeUnixMs)
	if resp.Err() != nil {
		if errors.Is(resp.Err(), dicedb.Nil) {
			// Default to current time
			cleanupTime := strconv.FormatInt(time.Now().UnixMilli(), 10)
			slog.Debug("Defaulting last cron cleanup time key since not set", slog.Any("cleanupTime", cleanupTime))
			resp := c.diceDBAdminClient.Client.Set(ctx, utils.LastCronCleanupTimeUnixMs, cleanupTime, -1)
			if resp.Err() != nil {
				slog.Error("Failed to set default value for last cron cleanup time key",
					slog.Any("err", resp.Err().Error()))
			}
		} else {
			slog.Error("Failed to get last cron cleanup time", slog.Any("err", resp.Err().Error()))
		}
	}

	for {
		select {
		case <-ticker.C:
			c.runCronTasks()
		case <-ctx.Done():
			slog.Info("Shutting down cleanup manager")
			return
		}
	}
}

func (c *CleanupManager) runCronTasks() {
	// Flush the user DiceDB instance
	resp := c.diceDBClient.Client.FlushDB(c.diceDBClient.Ctx)
	if resp.Err() != nil {
		slog.Error("Failed to flush keys from DiceDB user instance.")
	}

	// Update last cron run time on DiceDB instance
	cleanupTime := strconv.FormatInt(time.Now().UnixMilli(), 10)
	resp = c.diceDBAdminClient.Client.Set(c.diceDBClient.Ctx, utils.LastCronCleanupTimeUnixMs,
		cleanupTime, -1)
	slog.Debug("Updating last cron cleanup time key", slog.Any("cleanupTime", cleanupTime))
	if resp.Err() != nil {
		slog.Error("Failed to set LastCronCleanupTimeUnixMs")
	}
}
