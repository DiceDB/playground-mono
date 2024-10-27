package cleanup_cron

import (
	"context"
	"fmt"
	"net"
	"server/config"
	"server/internal/db"
	"server/internal/server"
	"server/internal/server/utils"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	setup_test "server/internal/tests/integration/setup"
)

// getDiceDBClient initializes the DiceDB client using the container's exposed port
func getDiceDBClient(container *setup_test.DiceDBContainer, config *config.Config, isAdmin bool) (*db.DiceDB, error) {
	// Get the container's host and port
	host, err := container.Container.Host(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get host: %v", err)
	}

	port, err := container.Container.MappedPort(context.Background(), "7379/tcp")
	if err != nil {
		return nil, fmt.Errorf("failed to get mapped port: %v", err)
	}

	address := net.JoinHostPort(host, port.Port())
	// the port is dynamically assigned so that's why
	// changing addr of diceDb and diceDbAdmin
	if isAdmin == true {
		config.DiceDBAdmin.Addr = address
	} else {
		config.DiceDB.Addr = address
	}

	client, err := db.InitDiceClient(config, isAdmin)
	if err != nil {
		return nil, fmt.Errorf("failed to create DiceDB client: %v", err)
	}

	return client, nil
}

func TestStartCleanupCronWithDiceDB(t *testing.T) {
	ctx := context.Background()

	configValue := config.LoadConfig()

	diceDbContainer, err := setup_test.InitializeDiceDBContainer(ctx)
	assert.NoError(t, err, "should initialize  DiceDB container")
	defer diceDbContainer.Cleanup(ctx)

	diceDbAdminContainer, err := setup_test.InitializeDiceDBContainer(ctx)
	assert.NoError(t, err, "should initialize Admin DiceDB container")
	defer diceDbAdminContainer.Cleanup(ctx)

	// Get userDemoClient and sysDiceClient from the running container
	diceDb, err := getDiceDBClient(diceDbContainer, configValue, false)
	assert.NoError(t, err, "should create DiceDB client")

	diceDbAdmin, err := getDiceDBClient(diceDbAdminContainer, configValue, true)
	assert.NoError(t, err, "should create Admin DiceDB client")

	// Add a sample key to test cleanup
	sampleValueResp := diceDb.Client.Set(ctx, "sample", "dummy", -1)
	assert.NoError(t, sampleValueResp.Err(), "should set sample key in DiceDB")

	wg := sync.WaitGroup{}
	// Register a cleanup manager, this runs user DiceDB instance cleanup job at configured frequency
	// setting a frequency for a second so that I can easily test it.
	cleanupManager := server.NewCleanupManager(diceDbAdmin, diceDb, time.Second)
	wg.Add(1)
	go cleanupManager.Run(ctx, &wg)

	time.Sleep(2 * time.Second)

	// Check if any keys are still present in diceDb (should be cleaned up)
	response := diceDb.Client.Keys(ctx, "*")
	assert.NoError(t, response.Err(), "should execute Keys command on DiceDB")
	assert.Equal(t, response.Val(), []string{}, "should have cleaned up keys in diceDb")

	// Retrieve the last cleanup time from diceDbAdmin
	cleanupTimeResp := diceDbAdmin.Client.Get(ctx, utils.LastCronCleanupTimeUnixMs)
	assert.NoError(t, cleanupTimeResp.Err(), "should execute GET command on adminDiceDb")
	assert.NotEmpty(t, cleanupTimeResp.Val(), "should have set playground_mono:last_cron_cleanup_run_time_unix_ms in diceDbAdmin")

}
