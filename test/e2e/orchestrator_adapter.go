package e2e

import (
	"context"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/stretchr/testify/mock"
	"testing"
)

type MockOrchestratorAdapter struct {
	mock *mock.Mock
}

func NewMockDockerAdapter() *MockOrchestratorAdapter {
	return &MockOrchestratorAdapter{
		mock: &mock.Mock{},
	}
}

func (d *MockOrchestratorAdapter) PullImage(ctx context.Context, imageName string) error {
	d.mock.MethodCalled("PullImage", ctx, imageName)
	return nil
}

func (d *MockOrchestratorAdapter) RunContainer(ctx context.Context, containerName string, config interface{}) (string, error) {
	d.mock.MethodCalled("RunContainer", ctx, containerName, config)
	return "fake-container-" + containerName, nil
}

func (d *MockOrchestratorAdapter) RemoveContainer(ctx context.Context, containerName string) error {
	d.mock.MethodCalled("RemoveContainer", ctx, containerName)
	return nil
}

func (d *MockOrchestratorAdapter) StoreConfiguration(ctx context.Context, containerName string, root string, config *adapter.AppConfig) (interface{}, error) {
	d.mock.MethodCalled("StoreConfiguration", ctx, containerName, root, config)
	return nil, nil
}

func (d *MockOrchestratorAdapter) GetContainerConfiguration(imageName string, containerName string, root string, httpPort int, gossipPort int, storedConfig interface{}) interface{} {
	d.mock.MethodCalled("GetContainerConfiguration", imageName, containerName, root, httpPort, gossipPort)
	return nil
}

func (d *MockOrchestratorAdapter) VerifyMocks(t *testing.T) {
	d.mock.AssertExpectations(t)
}
