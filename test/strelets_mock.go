package test

import (
	"context"
	"github.com/orbs-network/boyarin/strelets"
	"github.com/stretchr/testify/mock"
	"testing"
)

type StreletsMock struct {
	mock.Mock
}

func (s *StreletsMock) ProvisionVirtualChain(ctx context.Context, input *strelets.ProvisionVirtualChainInput) error {
	result := s.MethodCalled("ProvisionVirtualChain", ctx, input)
	return result.Error(0)
}

func (s *StreletsMock) RemoveVirtualChain(ctx context.Context, input *strelets.RemoveVirtualChainInput) error {
	result := s.MethodCalled("RemoveVirtualChain", ctx, input)
	return result.Error(0)
}

func (s *StreletsMock) VerifyMocks(t *testing.T) {
	s.AssertExpectations(t)
}

func (s *StreletsMock) UpdateReverseProxy(ctx context.Context, input *strelets.UpdateReverseProxyInput) error {
	result := s.MethodCalled("UpdateReverseProxy", input)
	return result.Error(0)
}
