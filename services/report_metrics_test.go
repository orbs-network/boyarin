package services

import (
	"fmt"
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

func TestInitializeMetrics(t *testing.T) {
	registry := prometheus.NewRegistry()

	metrics, err := InitializeMetrics(registry, []string{"disk0", "disk1"})
	require.NoError(t, err)
	require.NotNil(t, metrics)

	serializedMetrics, err := GetSerializedMetrics(registry)
	require.NoError(t, err)
	fmt.Println(serializedMetrics)
}
