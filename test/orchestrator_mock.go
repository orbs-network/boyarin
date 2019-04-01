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

func (a *OrchestratorMock) Prepare(ctx context.Context, serviceConfig *adapter.ServiceConfig, appConfig *adapter.AppConfig) (adapter.Runner, error) {
	res := a.MethodCalled("Prepare", ctx, serviceConfig, appConfig)
	return res.Get(0).(adapter.Runner), res.Error(1)
}

func (a *OrchestratorMock) RemoveContainer(ctx context.Context, containerName string) error {
	res := a.MethodCalled("RemoveContainer", ctx, containerName)
	return res.Error(1)
}

func (a *OrchestratorMock) PrepareReverseProxy(ctx context.Context, config string) (adapter.Runner, error) {
	res := a.MethodCalled("PrepareReverseProxy", ctx, config)
	return res.Get(0).(adapter.Runner), res.Error(1)
}

func (a *OrchestratorMock) Close() error {
	panic("not implemented")
	return nil
}

func (a *OrchestratorMock) GetStatus(ctx context.Context) (results []*adapter.ContainerStatus, err error) {
	panic("not implemented")
	return nil, nil
}

func NewOrchestratorAndRunnerMocks() (orchestrator *OrchestratorMock, virtualChainRunner *RunnerMock, reverseProxyRunner *RunnerMock) {
	orchestrator = &OrchestratorMock{}

	virtualChainRunner = &RunnerMock{}
	virtualChainRunner.On("Run", mock.Anything).Return(nil)

	reverseProxyRunner = &RunnerMock{}
	reverseProxyRunner.On("Run", mock.Anything).Return(nil)

	orchestrator.On("Prepare", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(virtualChainRunner, nil)
	orchestrator.On("PrepareReverseProxy", mock.Anything, mock.Anything).Return(reverseProxyRunner, nil)

	return
}
