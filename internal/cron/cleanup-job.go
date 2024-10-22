package cron

import (
	"context"
	"fmt"
	"log/slog"
	"server/internal/db"
	"server/util/cmds"
	"strconv"
	"time"
)

// starts a cron job that runs at the configured frequency
// and clean all the keys from use demo dicedb instance
func StartCleanupCron(ctx context.Context, userDemoClient *db.DiceDB, sysDiceClient *db.DiceDB, frequency int64) {
	// Convert seconds to time.Duration
	duration := time.Duration(frequency) * time.Second
	ticker := time.NewTicker(duration)

	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			slog.Info("Running cron job to clean up DiceDB keys")
			if err := cleanUpKeys(userDemoClient); err != nil {
				slog.Error("Failed to clean up keys", slog.Any("err", err))
			} else {

				if err := setLastCleanUpTime(sysDiceClient); err != nil {
					slog.Error("Failed inserting cleanup time", slog.Any("err", err))
				}

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

func setLastCleanUpTime(client *db.DiceDB) error {
	setCommand := &cmds.CommandRequest{
		Cmd:  "SET",
		Args: []string{"last-cleanup-time", getCurrentUnixTimestamp()},
	}

	setStatus, err := client.ExecuteCommand(setCommand)

	if err != nil {
		return err
	}

	if setStatus == "OK" {
		return nil
	} else {
		return fmt.Errorf("inappropriate response : %v ", setStatus)
	}

}

func getCurrentUnixTimestamp() string {
	return strconv.FormatInt(time.Now().Unix(), 10) // Returns Unix timestamp in seconds
}
