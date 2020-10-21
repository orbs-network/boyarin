package services

import (
	"bytes"
	"context"
	"fmt"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/orbs-network/boyarin/utils"
	"github.com/orbs-network/scribe/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/common/expfmt"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
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

type DiskMetric struct {
	Mountpoint  string
	TotalMbytes float64
	UsedMbytes  float64
	UsedPercent float64
}

type Metrics struct {
	CPULoadPercent    float64
	MemoryUsedPercent float64
	MemoryUsedMBytes  float64
	MemoryTotalMBytes float64
	EFSAccessTimeMs   uint64
	Disks             []DiskMetric
}

type PrometheusMetrics struct {
	cpuLoadPercent    prometheus.Gauge
	memoryUsedPercent prometheus.Gauge
	memoryUsedMbytes  prometheus.Gauge
	memoryTotalMbytes prometheus.Gauge
	efsAccessTimeMs   prometheus.Gauge
}

func InitializeAndUpdatePrometheusMetrics(registry *prometheus.Registry, metrics Metrics) PrometheusMetrics {
	prometheusMetrics := PrometheusMetrics{
		memoryTotalMbytes: promauto.With(registry).NewGauge(prometheus.GaugeOpts{
			Name: "memory_total_mbs",
		}),
		memoryUsedMbytes: promauto.With(registry).NewGauge(prometheus.GaugeOpts{
			Name: "memory_used_mbs",
		}),
		memoryUsedPercent: promauto.With(registry).NewGauge(prometheus.GaugeOpts{
			Name: "memory_used_percent",
		}),
		cpuLoadPercent: promauto.With(registry).NewGauge(prometheus.GaugeOpts{
			Name: "cpu_load_percent",
		}),
		efsAccessTimeMs: promauto.With(registry).NewGauge(prometheus.GaugeOpts{
			Name: "efs_access_time_ms",
		}),
	}

	prometheusMetrics.memoryTotalMbytes.Set(metrics.MemoryTotalMBytes)
	prometheusMetrics.memoryUsedMbytes.Set(float64(metrics.MemoryUsedMBytes))
	prometheusMetrics.memoryUsedPercent.Set(metrics.MemoryUsedPercent)
	prometheusMetrics.cpuLoadPercent.Set(metrics.CPULoadPercent)
	prometheusMetrics.efsAccessTimeMs.Set(float64(metrics.EFSAccessTimeMs))

	for _, diskMetric := range metrics.Disks {
		diskTotalMbs := promauto.With(registry).NewGauge(prometheus.GaugeOpts{
			Name: "disk_total_mbs",
			ConstLabels: map[string]string{
				"mountpoint": diskMetric.Mountpoint,
			},
		})

		diskUsedMbs := promauto.With(registry).NewGauge(prometheus.GaugeOpts{
			Name: "disk_used_mbs",
			ConstLabels: map[string]string{
				"mountpoint": diskMetric.Mountpoint,
			},
		})

		diskUsedPercent := promauto.With(registry).NewGauge(prometheus.GaugeOpts{
			Name: "disk_used_percent",
			ConstLabels: map[string]string{
				"mountpoint": diskMetric.Mountpoint,
			},
		})

		diskTotalMbs.Set(diskMetric.TotalMbytes)
		diskUsedMbs.Set(diskMetric.UsedMbytes)
		diskUsedPercent.Set(diskMetric.UsedPercent)
	}

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

	if cpuPercents, err := cpu.Percent(1*time.Second, false); err != nil {
		errors = append(errors, fmt.Errorf("failed to read cpu info: %s", err))
		logger.Error("failed to read cpu info", log.Error(err))
	} else {
		metrics.CPULoadPercent = cpuPercents[0]
	}

	if memInfo, err := mem.VirtualMemory(); err != nil {
		errors = append(errors, fmt.Errorf("failed to read memory info: %s", err))
		logger.Error("failed to read memory info", log.Error(err))
	} else {
		metrics.MemoryTotalMBytes = float64(memInfo.Total) / 1000 / 1000
		metrics.MemoryUsedMBytes = float64(memInfo.Used) / 1000 / 1000
		metrics.MemoryUsedPercent = memInfo.UsedPercent
	}

	diskMetrics, diskErrors := ReadDiskMetrics(logger)
	errors = append(errors, diskErrors...)
	metrics.Disks = diskMetrics

	accessTime, err := measureEFSAccessTime(ctx)
	if err != nil {
		errors = append(errors, fmt.Errorf("failed to measure EFS access time: %s", err))
		logger.Error("failed to measure EFS access time", log.Error(err))
	}
	metrics.EFSAccessTimeMs = accessTime

	return metrics, utils.AggregateErrors(errors)
}
