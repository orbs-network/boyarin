package test

import (
	"context"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/stretchr/testify/mock"
	"time"
)

type OrchestratorMock struct {
	mock.Mock
}

func (a *OrchestratorMock) PullImage(ctx context.Context, imageName string) error {
	panic("not implemented")
	return nil
}

func (a *OrchestratorMock) RunVirtualChain(ctx context.Context, serviceConfig *adapter.ServiceConfig, appConfig *adapter.AppConfig) error {
	res := a.MethodCalled("RunVirtualChain", ctx, serviceConfig, appConfig)
	return res.Error(0)
}

func (a *OrchestratorMock) ServiceRemove(ctx context.Context, containerName string) error {
	res := a.MethodCalled("ServiceRemove", ctx, containerName)
	return res.Error(0)
}

func (a *OrchestratorMock) RunReverseProxy(ctx context.Context, config *adapter.ReverseProxyConfig) error {
	res := a.MethodCalled("RunReverseProxy", ctx, config)
	return res.Error(0)
}

func (a *OrchestratorMock) Close() error {
	res := a.MethodCalled("Close")
	return res.Error(0)
}

func (a *OrchestratorMock) GetStatus(ctx context.Context, since time.Duration) (results []*adapter.ContainerStatus, err error) {
	res := a.MethodCalled("GetStatus", ctx, since)
	return res.Get(0).([]*adapter.ContainerStatus), res.Error(1)
}

func (a *OrchestratorMock) RunService(ctx context.Context, serviceConfig *adapter.ServiceConfig, appConfig *adapter.AppConfig) error {
	res := a.MethodCalled("RunService", ctx, serviceConfig, appConfig)
	return res.Error(0)
}

func (a *OrchestratorMock) GetOverlayNetwork(ctx context.Context, name string) (string, error) {
	res := a.MethodCalled("GetOverlayNetwork", ctx, name)
	return res.String(0), res.Error(1)
}
