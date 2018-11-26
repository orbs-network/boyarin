package e2e

import (
	"context"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/stretchr/testify/mock"
	"testing"
)

type MockOrchestratorAdapter struct {
	mock   *mock.Mock
	runner *MockRunner
}

type MockRunner struct {
	mock *mock.Mock
}

func NewMockDockerAdapter() *MockOrchestratorAdapter {
	return &MockOrchestratorAdapter{
		mock: &mock.Mock{},
		runner: &MockRunner{
			mock: &mock.Mock{},
		},
	}
}

func (d *MockOrchestratorAdapter) PullImage(ctx context.Context, imageName string) error {
	d.mock.MethodCalled("PullImage", ctx, imageName)
	return nil
}

func (d *MockOrchestratorAdapter) Prepare(ctx context.Context, imageName string, containerName string, httpPort int, gossipPort int, config *adapter.AppConfig) (adapter.Runner, error) {
	d.mock.MethodCalled("Prepare", ctx, containerName, config)
	return d.runner, nil
}

func (d *MockOrchestratorAdapter) RemoveContainer(ctx context.Context, containerName string) error {
	d.mock.MethodCalled("RemoveContainer", ctx, containerName)
	return nil
}

func (d *MockOrchestratorAdapter) VerifyMocks(t *testing.T) {
	d.mock.AssertExpectations(t)
	d.runner.mock.AssertExpectations(t)
}

func (r *MockRunner) Run(ctx context.Context) error {
	r.mock.MethodCalled("Run", ctx)
	return nil
}

func (d *MockOrchestratorAdapter) UpdateReverseProxy(ctx context.Context, config string) error {
	panic("not implemented")
	return nil
}
