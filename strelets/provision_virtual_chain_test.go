package strelets

import (
	"context"
	"fmt"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

type mockOrchestrator struct {
	mock.Mock
}

func (a *mockOrchestrator) PullImage(ctx context.Context, imageName string) error {
	panic("not implemented")
	return nil
}

func (a *mockOrchestrator) Prepare(ctx context.Context, imageName string, containerName string, httpPort int, gossipPort int, config *adapter.AppConfig) (adapter.Runner, error) {
	res := a.MethodCalled("Prepare", ctx, imageName, containerName, httpPort, gossipPort, config)
	return res.Get(0).(adapter.Runner), res.Error(1)
}

func (a *mockOrchestrator) RemoveContainer(ctx context.Context, containerName string) error {
	panic("not implemented")
	return nil
}

func (a *mockOrchestrator) PrepareReverseProxy(ctx context.Context, config string) (adapter.Runner, error) {
	panic("not implemented")
	return nil, nil
}

func (a *mockOrchestrator) Close() error {
	panic("not implemented")
	return nil
}

// Set failedAttempts to sabotage Run()
type mockRunner struct {
	mock.Mock
	attempts       int
	failedAttempts int
}

func (m *mockRunner) Run(ctx context.Context) (err error) {
	m.MethodCalled("Run", ctx)

	if m.attempts == m.failedAttempts {
		return
	}
	m.attempts += 1

	return fmt.Errorf("some error")
}

func TestStrelets_ProvisionVirtualChain(t *testing.T) {
	orchestrator := &mockOrchestrator{}
	s := NewStrelets(orchestrator)

	runner := &mockRunner{failedAttempts: 0}
	runner.On("Run", mock.Anything).Return(nil)

	orchestrator.On("Prepare", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(runner, nil)

	peers := make(PeersMap)
	err := s.ProvisionVirtualChain(context.Background(), &ProvisionVirtualChainInput{
		VirtualChain: &VirtualChain{
			Id: VirtualChainId(1972),
		},
		Peers:             &peers,
		KeyPairConfigPath: "./fixtures/keys.json",
	})

	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "Prepare", 1)
	runner.AssertNumberOfCalls(t, "Run", 1)
}

func TestStrelets_ProvisionVirtualChainWithRetries(t *testing.T) {
	orchestrator := &mockOrchestrator{}
	s := NewStrelets(orchestrator)

	runner := &mockRunner{failedAttempts: 2}
	runner.On("Run", mock.Anything).Return(nil)

	orchestrator.On("Prepare", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(runner, nil)

	peers := make(PeersMap)
	err := s.ProvisionVirtualChain(context.Background(), &ProvisionVirtualChainInput{
		VirtualChain: &VirtualChain{
			Id: VirtualChainId(1972),
		},
		Peers:             &peers,
		KeyPairConfigPath: "./fixtures/keys.json",
	})

	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "Prepare", 3)
	runner.AssertNumberOfCalls(t, "Run", 3)
}
