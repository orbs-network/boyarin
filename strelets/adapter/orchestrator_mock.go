package adapter

import (
	"context"
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

func (a *OrchestratorMock) RunVirtualChain(ctx context.Context, serviceConfig *ServiceConfig, appConfig *AppConfig) error {
	res := a.MethodCalled("RunVirtualChain", ctx, serviceConfig, appConfig)
	return res.Error(0)
}

func (a *OrchestratorMock) RemoveService(ctx context.Context, containerName string) error {
	res := a.MethodCalled("RemoveService", ctx, containerName)
	return res.Error(0)
}

func (a *OrchestratorMock) RunReverseProxy(ctx context.Context, config *ReverseProxyConfig) error {
	res := a.MethodCalled("RunReverseProxy", ctx, config)
	return res.Error(0)
}

func (a *OrchestratorMock) Close() error {
	res := a.MethodCalled("Close")
	return res.Error(0)
}

func (a *OrchestratorMock) GetStatus(ctx context.Context, since time.Duration) (results []*ContainerStatus, err error) {
	res := a.MethodCalled("GetStatus", ctx, since)
	return res.Get(0).([]*ContainerStatus), res.Error(1)
}

func (a *OrchestratorMock) RunService(ctx context.Context, serviceConfig *ServiceConfig, appConfig *AppConfig) error {
	res := a.MethodCalled("RunService", ctx, serviceConfig, appConfig)
	return res.Error(0)
}

func (a *OrchestratorMock) GetOverlayNetwork(ctx context.Context, name string) (string, error) {
	res := a.MethodCalled("GetOverlayNetwork", ctx, name)
	return res.String(0), res.Error(1)
}

func (a *OrchestratorMock) PurgeServiceData(ctx context.Context, containerName string) error {
	res := a.MethodCalled("PurgeServiceData", ctx, containerName)
	return res.Error(1)
}

func (a *OrchestratorMock) PurgeVirtualChainData(ctx context.Context, nodeAddress string, vcId uint32, containerName string) error {
	res := a.MethodCalled("PurgeVirtualChainData", ctx, nodeAddress, vcId, containerName)
	return res.Error(1)
}
