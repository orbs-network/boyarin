package test

import (
	"context"
	"github.com/orbs-network/boyarin/strelets"
	"github.com/stretchr/testify/mock"
	"testing"
)

type streletsMock struct {
	mock *mock.Mock
}

func (s *streletsMock) ProvisionVirtualChain(ctx context.Context, input *strelets.ProvisionVirtualChainInput) error {
	s.mock.MethodCalled("ProvisionVirtualChain", ctx, input)
	return nil
}

func (s *streletsMock) RemoveVirtualChain(ctx context.Context, input *strelets.RemoveVirtualChainInput) error {
	s.mock.MethodCalled("RemoveVirtualChain", input)
	return nil
}

func (s *streletsMock) VerifyMocks(t *testing.T) {
	s.mock.AssertExpectations(t)
}
