package setup_test

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type DiceDBContainer struct {
	Container testcontainers.Container
}

func InitializeDiceDBContainer(ctx context.Context) (*DiceDBContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "dicedb/dicedb:latest",
		ExposedPorts: []string{"7379/tcp"},
		WaitingFor:   wait.ForLog("starting DiceDB"),
	}

	diceDBContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start diceDB container: %v", err)
	}

	return &DiceDBContainer{
		Container: diceDBContainer,
	}, nil
}

func (dbc *DiceDBContainer) Cleanup(ctx context.Context) error {
	err := dbc.Container.Terminate(ctx)
	if err != nil {
		return fmt.Errorf("could not terminate container: %v", err)
	}
	return nil
}
