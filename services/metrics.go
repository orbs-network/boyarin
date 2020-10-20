package services

import (
	"bytes"
	"context"
	"fmt"
	"github.com/c9s/goprocinfo/linux"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/orbs-network/boyarin/utils"
	"github.com/orbs-network/scribe/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/common/expfmt"
	"os"
	"path/filepath"
	"time"
)

func GetSerializedMetrics(registry *prometheus.Registry) (value string, err error) {
	mfs, err := registry.Gather()
	if err != nil {
		return
	}

	w := bytes.NewBufferString("")
	enc := expfmt.NewEncoder(w, expfmt.FmtText)

	for _, mf := range mfs {
		if err = enc.Encode(mf); err != nil {
			return
		}
	}

	return w.String(), nil
}

type Metrics struct {
	CPULoad          uint64
	MemoryLoad       uint64
	UsedMemoryMBytes uint64
	EFSAccessTimeMs  uint64
}

type PrometheusMetrics struct {
	cpuLoad          prometheus.Gauge
	memoryLoad       prometheus.Gauge
	usedMemoryMBytes prometheus.Gauge
	efsAccessTimeMs  prometheus.Gauge
}

func getCPULoad() (uint64, error) {
	cpuStats, err := linux.ReadStat("/proc/stat")
	if err != nil {
		return 0, err
	}
	cpu := cpuStats.CPUStatAll
	return cpu.User + cpu.Nice + cpu.System + cpu.Idle, nil
}

func InitializeAndUpdatePrometheusMetrics(registry *prometheus.Registry, metrics Metrics) PrometheusMetrics {
	prometheusMetrics := PrometheusMetrics{
		usedMemoryMBytes: promauto.With(registry).NewGauge(prometheus.GaugeOpts{
			Name: "used_memory_mbs",
			Help: "total memory used in megabytes",
		}),
		memoryLoad: promauto.With(registry).NewGauge(prometheus.GaugeOpts{
			Name: "memory_load_percent",
			Help: "total memory load in percent",
		}),
		cpuLoad: promauto.With(registry).NewGauge(prometheus.GaugeOpts{
			Name: "cpu_load_percent",
			Help: "CPU load across all processors in percent",
		}),
		efsAccessTimeMs: promauto.With(registry).NewGauge(prometheus.GaugeOpts{
			Name: "efs_access_time_ms",
			Help: "EFS access time in milliseconds",
		}),
	}

	prometheusMetrics.usedMemoryMBytes.Set(float64(metrics.UsedMemoryMBytes))
	prometheusMetrics.cpuLoad.Set(float64(metrics.CPULoad))
	prometheusMetrics.memoryLoad.Set(float64(metrics.MemoryLoad))
	prometheusMetrics.efsAccessTimeMs.Set(float64(metrics.EFSAccessTimeMs))

	return prometheusMetrics
}

func measureEFSAccessTime(ctx context.Context) (uint64, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
		start := time.Now()
		err := filepath.Walk(adapter.DEFAULT_EFS_PATH, func(path string, info os.FileInfo, err error) error {
			return err
		})

		return uint64(time.Since(start).Milliseconds()), err
	}
}

func CollectMetrics(ctx context.Context, logger log.Logger) (metrics Metrics, aggregateError error) {
	var errors []error

	cpu0, errCPU0 := getCPULoad()

	<-time.After(time.Second)

	cpu1, errCPU1 := getCPULoad()
	if errCPU0 != nil {
		errors = append(errors, fmt.Errorf("failed to read /proc/stat: %s", errCPU0))
		logger.Error("failed to read /proc/stat", log.Error(errCPU0))
	} else if errCPU1 != nil {
		errors = append(errors, fmt.Errorf("failed to read /proc/stat: %s", errCPU1))
		logger.Error("failed to read /proc/stat", log.Error(errCPU1))
	} else {
		metrics.CPULoad = (cpu1 - cpu0) * 100
	}

	memInfo, err := linux.ReadMemInfo("/proc/meminfo")
	if err != nil {
		errors = append(errors, fmt.Errorf("failed to read /proc/meminfo: %s", err))
		logger.Error("failed to read /proc/meminfo", log.Error(err))
	} else {
		metrics.UsedMemoryMBytes = (memInfo.MemAvailable - memInfo.MemFree) / 1000
		metrics.MemoryLoad = (memInfo.MemAvailable - memInfo.MemFree) / memInfo.MemAvailable * 100
	}

	accessTime, err := measureEFSAccessTime(ctx)
	if err != nil {
		errors = append(errors, fmt.Errorf("failed to measure EFS access time: %s", err))
		logger.Error("failed to measure EFS access time", log.Error(err))
	}
	metrics.EFSAccessTimeMs = accessTime

	return metrics, utils.AggregateErrors(errors)
}
