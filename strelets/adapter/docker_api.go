package adapter

import (
	"context"
)

type DockerAPI interface {
	PullImage(ctx context.Context, imageName string) error
	RunContainer(ctx context.Context, containerName string, config map[string]interface{}) (string, error)
	RemoveContainer(ctx context.Context, containerName string) error
}
