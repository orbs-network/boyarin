package services

import (
	"context"
	"fmt"
	"github.com/orbs-network/scribe/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetSerializedMetrics(t *testing.T) {
	registry := prometheus.NewRegistry()

	opsProcessed := promauto.With(registry).NewCounter(prometheus.CounterOpts{
		Name: "myapp_processed_ops_total",
		Help: "The total number of processed events",
	})

	for i := 0; i < 10; i++ {
		opsProcessed.Inc()
	}

	serializedMetrics, err := GetSerializedMetrics(registry)
	require.NoError(t, err)
	require.Contains(t, serializedMetrics, "myapp_processed_ops_total 10")
}

func TestCollectMetrics(t *testing.T) {
	registry := prometheus.NewRegistry()

	logger := log.GetLogger()
	//ctx, cancel := context.WithTimeout(context.Background(), 0)
	//defer cancel()
	ctx := context.Background()

	m, _ := CollectMetrics(ctx, logger)
	metrics := InitializeAndUpdatePrometheusMetrics(registry, m)
	require.NotNil(t, metrics)

	serializedMetrics, err := GetSerializedMetrics(registry)
	require.NoError(t, err)
	fmt.Println(serializedMetrics)
}
