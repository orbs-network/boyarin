package test

import (
	"context"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/stretchr/testify/mock"
)

type OrchestratorMock struct {
	mock.Mock
}

func (a *OrchestratorMock) PullImage(ctx context.Context, imageName string) error {
	panic("not implemented")
	return nil
}

func (a *OrchestratorMock) Prepare(ctx context.Context, imageName string, containerName string, httpPort int, gossipPort int, config *adapter.AppConfig) (adapter.Runner, error) {
	res := a.MethodCalled("Prepare", ctx, imageName, containerName, httpPort, gossipPort, config)
	return res.Get(0).(adapter.Runner), res.Error(1)
}

func (a *OrchestratorMock) RemoveContainer(ctx context.Context, containerName string) error {
	panic("not implemented")
	return nil
}

func (a *OrchestratorMock) PrepareReverseProxy(ctx context.Context, config string) (adapter.Runner, error) {
	panic("not implemented")
	return nil, nil
}

func (a *OrchestratorMock) Close() error {
	panic("not implemented")
	return nil
}
