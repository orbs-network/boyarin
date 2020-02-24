package boyar

//func TestBoyar_ProvisionHttpAPIEndpointWithNoConfigChanges(t *testing.T) {
//	cfg := getJSONConfig(t, ConfigWithActiveVchains)
//
//	orchestrator := &adapter.OrchestratorMock{}
//	orchestrator.On("RunReverseProxy", mock.Anything, mock.Anything).Return(nil).Once()
//
//	cache := NewCache()
//	b := NewBoyar(orchestrator, cfg, cache, helpers.DefaultTestLogger())
//
//	err := b.ProvisionHttpAPIEndpoint(context.Background())
//	require.NoError(t, err)
//	orchestrator.AssertExpectations(t)
//
//	nginxConfig := getNginxCompositeConfig(cfg)
//	assert.False(t, cache.nginx.CheckNewJsonValue(getNginxConfig(nginxConfig.Chains, nginxConfig.IP, false)))
//
//	err = b.ProvisionHttpAPIEndpoint(context.Background())
//	require.NoError(t, err)
//	orchestrator.AssertExpectations(t)
//}

//func TestBoyar_ProvisionHttpAPIEndpointReprovisionsIfConfigChanges(t *testing.T) {
//	cfg := getJSONConfig(t, ConfigWithActiveVchains)
//
//	orchestrator := &adapter.OrchestratorMock{}
//	orchestrator.On("RunReverseProxy", mock.Anything, mock.Anything).Return(nil).Once()
//
//	cache := NewCache()
//	b := NewBoyar(orchestrator, cfg, cache, helpers.DefaultTestLogger())
//
//	err := b.ProvisionHttpAPIEndpoint(context.Background())
//	require.NoError(t, err)
//	orchestrator.AssertExpectations(t)
//
//	assert.False(t, cache.nginx.CheckNewJsonValue(getNginxConfig(cfg.Chains(), "127.0.0.1", false)))
//
//	orchestrator.On("RunReverseProxy", mock.Anything, mock.Anything).Return(nil).Once()
//	cfg.Chains()[0].HttpPort = 9125
//
//	err = b.ProvisionHttpAPIEndpoint(context.Background())
//	require.NoError(t, err)
//	orchestrator.AssertExpectations(t)
//}
