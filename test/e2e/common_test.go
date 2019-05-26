package e2e

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/orbs-network/boyarin/strelets"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

const HTTP_PORT = 8080
const GOSSIP_PORT = 4400

const WAIT_FOR_BLOCK_TIMEOUT = 3 * time.Minute

func getConfigPath() string {
	configPath := "../../e2e-config/"
	if configPathFromEnv := os.Getenv("E2E_CONFIG"); configPathFromEnv != "" {
		configPath = configPathFromEnv
	}

	return configPath
}

func removeAllDockerVolumes(t *testing.T) {
	t.Log("Removing all docker volumes")

	ctx := context.Background()
	client, err := client.NewClientWithOpts(client.WithVersion(adapter.DOCKER_API_VERSION))
	if err != nil {
		t.Errorf("could not connect to docker: %s", err)
		t.FailNow()
	}

	if containers, err := client.ContainerList(ctx, types.ContainerListOptions{}); err != nil {
		t.Errorf("could not list docker containers: %s", err)
		t.FailNow()
	} else {
		for _, c := range containers {
			t.Log("container", c.Names[0], "is still up with state", c.State)
		}
	}

	if volumes, err := client.VolumeList(ctx, filters.Args{}); err != nil {
		t.Errorf("could not list docker volumes: %s", err)
		t.FailNow()
	} else {
		for _, v := range volumes.Volumes {
			fmt.Println("removing volume:", v.Name)

			if err := strelets.Try(ctx, 10, 1*time.Second, 100*time.Millisecond, func(ctxWithTimeout context.Context) error {
				return client.VolumeRemove(ctx, v.Name, true)
			}); err != nil {
				t.Errorf("could not list docker volumes: %s", err)
				t.FailNow()
			}
		}
	}
}

func removeAllServices(t *testing.T) {
	t.Log("Removing all swarm services")

	ctx := context.Background()
	client, err := client.NewClientWithOpts(client.WithVersion(adapter.DOCKER_API_VERSION))
	if err != nil {
		t.Errorf("could not connect to docker: %s", err)
		t.FailNow()
	}

	services, err := client.ServiceList(ctx, types.ServiceListOptions{})
	require.NoError(t, err)

	for _, s := range services {
		client.ServiceRemove(ctx, s.ID)
	}

	// FIXME add Eventually() to wait until shutdown
}
