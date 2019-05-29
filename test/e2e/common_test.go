package e2e

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/strelets"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/orbs-network/boyarin/test/helpers"
	"github.com/stretchr/testify/require"
	"os"
	"strings"
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

func getKeyPairConfigForNode(i int, addressOnly bool) []byte {
	cfg, err := config.NewKeysConfig(fmt.Sprintf("%s/node%d/keys.json", getConfigPath(), i))
	if err != nil {
		panic(err)
	}

	return cfg.JSON(addressOnly)
}

func removeAllDockerVolumes(t *testing.T) {
	fmt.Println("Removing all docker volumes")

	ctx := context.Background()
	client, err := client.NewClientWithOpts(client.WithVersion(adapter.DOCKER_API_VERSION))
	if err != nil {
		t.Errorf("could not connect to docker: %s", err)
		t.FailNow()
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
				if containers, err := client.ContainerList(ctx, types.ContainerListOptions{
					All: true,
				}); err != nil {
					t.Errorf("could not list docker containers: %s", err)
					t.FailNow()
				} else {
					for _, c := range containers {
						t.Log("container", c.Names[0], "is still up with state", c.State)
					}
				}
				t.FailNow()
			}
		}
	}
}

func removeAllServices(t *testing.T) {
	fmt.Println("Removing all swarm services")

	ctx := context.Background()
	client, err := client.NewClientWithOpts(client.WithVersion(adapter.DOCKER_API_VERSION))
	if err != nil {
		t.Errorf("could not connect to docker: %s", err)
		t.FailNow()
	}

	services, err := client.ServiceList(ctx, types.ServiceListOptions{})
	require.NoError(t, err)

	for _, s := range services {
		fmt.Println("Removing service", s.Spec.Name)
		client.ServiceRemove(ctx, s.ID)
	}

	require.Truef(t, helpers.Eventually(20*time.Second, func() bool {
		services, err := client.ServiceList(ctx, types.ServiceListOptions{})
		if err != nil {
			return false
		}

		return len(services) == 0
	}), "failed to remove swarm services in time")

	// Depends on CI
	CI := os.Getenv("CI") != ""
	numContainers := 0
	if CI {
		numContainers = 2
	}

	containers, err := client.ContainerList(ctx, types.ContainerListOptions{
		All: true,
	})
	require.NoError(t, err)

	for _, c := range containers {
		if CI && (strings.Contains(c.Names[0], "ganache") || strings.Contains(c.Names[0], "boyar")) {
			continue
		}

		fmt.Println("removing container", c.Names[0])
		client.ContainerRemove(ctx, c.ID, types.ContainerRemoveOptions{
			Force: true,
		})
	}

	require.Truef(t, helpers.Eventually(20*time.Second, func() bool {
		containers, err := client.ContainerList(ctx, types.ContainerListOptions{
			All: true,
		})
		if err != nil {
			return false
		}

		return len(containers) <= numContainers
	}), "failed to remove docker containers in time")
}

func withCleanContext(t *testing.T, f func(t *testing.T)) {
	helpers.SkipUnlessSwarmIsEnabled(t)

	removeAllServices(t)
	removeAllDockerVolumes(t)

	t.Run("clean docker context", f)

	removeAllServices(t)
	removeAllDockerVolumes(t)
}
