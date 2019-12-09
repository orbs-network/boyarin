package helpers

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	dockerClient "github.com/docker/docker/client"
	"github.com/orbs-network/boyarin/utils"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

const DOCKER_API_VERSION = "1.40"

func removeAllDockerVolumes(t *testing.T) {
	fmt.Println("Removing all docker volumes")

	ctx := context.Background()
	client, err := dockerClient.NewClientWithOpts(dockerClient.WithVersion(DOCKER_API_VERSION))
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

			if err := utils.Try(ctx, 10, 1*time.Second, 100*time.Millisecond, func(ctxWithTimeout context.Context) error {
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
	client, err := dockerClient.NewClientWithOpts(dockerClient.WithVersion(DOCKER_API_VERSION))
	if err != nil {
		t.Errorf("could not connect to docker: %s", err)
		t.FailNow()
	}

	services, err := client.ServiceList(ctx, types.ServiceListOptions{})
	require.NoError(t, err)

	for _, s := range services {
		fmt.Println("Removing service", s.Spec.Name)
		err = client.ServiceRemove(ctx, s.ID)
		if err != nil {
			fmt.Printf("error removing service '%s': %p\n", s.Spec.Name, err)
		}
	}

	require.Truef(t, Eventually(20*time.Second, func() bool {
		services, err := client.ServiceList(ctx, types.ServiceListOptions{})
		if err != nil {
			return false
		}

		return len(services) == 0
	}), "failed to remove swarm services in time")

	containers, err := client.ContainerList(ctx, types.ContainerListOptions{
		All: true,
	})
	require.NoError(t, err)

	for _, c := range containers {
		fmt.Println("removing container", c.Names[0])
		err = client.ContainerRemove(ctx, c.ID, types.ContainerRemoveOptions{
			Force: true,
		})
		if err != nil {
			fmt.Printf("error removing container '%s': %p", c.Names[0], err)
		}
	}

	require.Truef(t, Eventually(20*time.Second, func() bool {
		containers, err := client.ContainerList(ctx, types.ContainerListOptions{
			All: true,
		})
		if err != nil {
			return false
		}

		return len(containers) <= 0
	}), "failed to remove docker containers in time")
}

func InitSwarmEnvironment(t *testing.T, ctx context.Context) {
	SkipUnlessSwarmIsEnabled(t)

	removeAllServices(t)
	removeAllDockerVolumes(t)
	fmt.Println("swarm cleared")

	LogSwarmServices(t, ctx)
}
