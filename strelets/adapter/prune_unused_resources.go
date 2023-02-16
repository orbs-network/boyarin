package adapter

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
)

func (d *dockerSwarmOrchestrator) PruneUnusedResources(ctx context.Context) error {

	fmt.Println("executing docker system prune all")

	client := d.client

	prune, err := client.BuildCachePrune(ctx, types.BuildCachePruneOptions{All: true})
	if err != nil {
		return err
	}
	fmt.Printf("pruned docker build cache: %v\n", len(prune.CachesDeleted))

	networksPrune, err := client.NetworksPrune(ctx, filters.Args{})
	if err != nil {
		return err
	}
	fmt.Printf("pruned docker networks: %v\n", len(networksPrune.NetworksDeleted))

	imagePrune, err := client.ImagesPrune(ctx, filters.Args{})
	if err != nil {
		return err
	}
	fmt.Printf("pruned docker images: %v\n", len(imagePrune.ImagesDeleted))

	volumesPrune, err := client.VolumesPrune(ctx, filters.Args{})
	if err != nil {
		return err
	}
	fmt.Printf("pruned docker volumes: %v\n", len(volumesPrune.VolumesDeleted))

	containersPrune, err := client.ContainersPrune(ctx, filters.Args{})
	if err != nil {
		return err
	}
	fmt.Printf("pruned docker containers: %v\n", len(containersPrune.ContainersDeleted))

	return nil

}
