package cron

import (
	"context"
	"fmt"
	"log/slog"
	"server/internal/db"
	"server/util/cmds"
	"time"
)

var LastCleanUpTime int64

// starts a cron job that runs at the configured frequency
// and clean all the keys from dicedb instance
func StartCleanupCron(ctx context.Context, client *db.DiceDB, frequency int64) {
	// Convert seconds to time.Duration
	duration := time.Duration(frequency) * time.Second
	ticker := time.NewTicker(duration)

	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			slog.Info("Running cron job to clean up DiceDB keys")
			if err := cleanUpKeys(client); err != nil {
				slog.Error("Failed to clean up keys", slog.Any("err", err))
			} else {
				LastCleanUpTime = getCurrentUnixTimestamp()
				slog.Info("Successfully cleaned up keys in DiceDB")
			}
		case <-ctx.Done():
			return
		}
	}
}

// cleanUpKeys performs the cleanup of stale or unused keys in DiceDB
func cleanUpKeys(client *db.DiceDB) error {
	flushCommand := &cmds.CommandRequest{
		Cmd: "FLUSHDB",
	}
	flushStatus, err := client.ExecuteCommand(flushCommand)

	if err != nil {
		return err
	}

	if flushStatus == "OK" {
		return nil
	} else {
		return fmt.Errorf("inappropriate response : %v ", flushStatus)
	}
}

func getCurrentUnixTimestamp() int64 {
	return time.Now().Unix() // Returns Unix timestamp in seconds
}
