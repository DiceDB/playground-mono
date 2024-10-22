package middleware_test

import (
	"context"
	"fmt"
	"net"
	"server/internal/cron"
	"server/internal/db"
	"server/util/cmds"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	setup_test "server/internal/tests/integration/setup"
)

// getDiceDBClient initializes the DiceDB client using the container's exposed port
func getDiceDBClient(container *setup_test.DiceDBContainer) (*db.DiceDB, error) {
	// Get the container's host and port
	host, err := container.Container.Host(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get host: %v", err)
	}

	port, err := container.Container.MappedPort(context.Background(), "7379/tcp")
	if err != nil {
		return nil, fmt.Errorf("failed to get mapped port: %v", err)
	}

	// Connect to the DiceDB instance running in the container
	address := net.JoinHostPort(host, port.Port())
	client, err := db.InitDiceClient(address) // You may need to adjust this based on your db.DiceDB client creation
	if err != nil {
		return nil, fmt.Errorf("failed to create DiceDB client: %v", err)
	}

	return client, nil
}

func TestStartCleanupCronWithDiceDB(t *testing.T) {
	ctx := context.Background()

	userDemodiceDBContainer, err := setup_test.InitializeDiceDBContainer(ctx)
	assert.NoError(t, err, "should initialize User Demo DiceDB container")
	defer userDemodiceDBContainer.Cleanup(ctx)

	sysDiceDBContainer, err := setup_test.InitializeDiceDBContainer(ctx)
	assert.NoError(t, err, "should initialize Sys DiceDB container")
	defer sysDiceDBContainer.Cleanup(ctx)

	// Get userDemoClient and sysDiceClient from the running container
	userDemoClient, err := getDiceDBClient(userDemodiceDBContainer)
	assert.NoError(t, err, "should create DiceDB client for userDemo")

	sysDiceClient, err := getDiceDBClient(sysDiceDBContainer)
	assert.NoError(t, err, "should create DiceDB client for sysDice")

	// Set up a cancellable context
	cleanupCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// For testing starting a cron job with 1 seconds frequency
	go cron.StartCleanupCron(cleanupCtx, userDemoClient, sysDiceClient, 1)

	// Wait for some time to allow the cron job to run
	time.Sleep(2 * time.Second)

	// Validate if the cleanup happened (you can check by sending commands to userDemoClient)
	flushCheckCmd := &cmds.CommandRequest{
		Cmd:  "GET",
		Args: []string{"*"},
	}

	response, err := userDemoClient.ExecuteCommand(flushCheckCmd)
	assert.NoError(t, err, "should execute INFO command on DiceDB")
	assert.Contains(t, response, "(nil)", "should have cleaned up keys in userDemoClient")

	// Validate if the cleanup time was set in sysDiceClient
	getLastCleanupCmd := &cmds.CommandRequest{
		Cmd:  "GET",
		Args: []string{"last-cleanup-time"},
	}

	cleanupTime, err := sysDiceClient.ExecuteCommand(getLastCleanupCmd)
	assert.NoError(t, err, "should execute GET command on sysDice")
	assert.NotEmpty(t, cleanupTime, "should have set last-cleanup-time in sysDiceClient")

}
