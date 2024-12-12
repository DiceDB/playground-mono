package commands

import (
	"context"
	"log"
	"os"
	setup_test "server/internal/tests/integration/setup"
	"testing"
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	diceDBC, err := setup_test.InitializeDiceDBContainer(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err := diceDBC.Cleanup(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}()

	// testcontainers-go maps container port to a random host port generated at runtime.
	// Hence, we get the mapped port and update the DICEDB_ADDR environment variable.
	port, err := diceDBC.Container.MappedPort(ctx, "7379")
	if err != nil {
		log.Fatalf("failed to get container port mapping: %v", err)
	}
	diceDBAddr := "localhost:" + port.Port()
	os.Setenv("DICEDB_ADDR", diceDBAddr)
	os.Setenv("ENVIRONMENT", "local")

	code := m.Run()
	os.Exit(code)
}
