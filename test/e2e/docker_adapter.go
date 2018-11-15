package e2e

import (
	"context"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/stretchr/testify/mock"
	"testing"
)

type MockDockerAdapter struct {
	mock *mock.Mock
}

func NewMockDockerAdapter() *MockDockerAdapter {
	return &MockDockerAdapter{
		mock: &mock.Mock{},
	}
}

func (d *MockDockerAdapter) PullImage(ctx context.Context, imageName string) error {
	d.mock.MethodCalled("PullImage", ctx, imageName)
	return nil
}

func (d *MockDockerAdapter) RunContainer(ctx context.Context, containerName string, config interface{}) (string, error) {
	d.mock.MethodCalled("RunContainer", ctx, containerName, config)
	return "fake-container-" + containerName, nil
}

func (d *MockDockerAdapter) RemoveContainer(ctx context.Context, containerName string) error {
	d.mock.MethodCalled("RemoveContainer", ctx, containerName)
	return nil
}

func (d *MockDockerAdapter) StoreConfiguration(ctx context.Context, containerName string, root string, config *adapter.AppConfig) error {
	d.mock.MethodCalled("StoreConfiguration", ctx, containerName, root, config)
	return nil
}

func (d *MockDockerAdapter) GetContainerConfiguration(imageName string, containerName string, root string, httpPort int, gossipPort int) interface{} {
	d.mock.MethodCalled("GetContainerConfiguration", imageName, containerName, root, httpPort, gossipPort)
	return nil
}

func (d *MockDockerAdapter) VerifyMocks(t *testing.T) {
	d.mock.AssertExpectations(t)
}
