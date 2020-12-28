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
	"github.com/shirou/gopsutil/process"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
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

type ProcessMetric struct {
	Name             string
	Command          string
	MemoryUsedMbytes float64
	PID              int32
	ParentPID        int32
}

type Metrics struct {
	BoyarUptime       float64
	CPULoadPercent    float64
	MemoryUsedPercent float64
	MemoryUsedMBytes  float64
	MemoryTotalMBytes float64
	EFSAccessTimeMs   uint64
	Disks             []DiskMetric
	Processes         []ProcessMetric
}

type PrometheusMetrics struct {
	cpuLoadPercent    prometheus.Gauge
	memoryUsedPercent prometheus.Gauge
	memoryUsedMbytes  prometheus.Gauge
	memoryTotalMbytes prometheus.Gauge
	efsAccessTimeMs   prometheus.Gauge
}

func InitializeAndUpdatePrometheusMetrics(registry *prometheus.Registry, metrics Metrics) {
	boyarUptimeSeconds := promauto.With(registry).NewGauge(prometheus.GaugeOpts{
		Name: "boyar_uptime_seconds",
	})

	boyarUptimeSeconds.Set(metrics.BoyarUptime)

	memoryTotalMbytes := promauto.With(registry).NewGauge(prometheus.GaugeOpts{
		Name: "memory_total_mbs",
	})
	memoryUsedMbytes := promauto.With(registry).NewGauge(prometheus.GaugeOpts{
		Name: "memory_used_mbs",
	})
	memoryUsedPercent := promauto.With(registry).NewGauge(prometheus.GaugeOpts{
		Name: "memory_used_percent",
	})
	cpuLoadPercent := promauto.With(registry).NewGauge(prometheus.GaugeOpts{
		Name: "cpu_load_percent",
	})
	efsAccessTimeMs := promauto.With(registry).NewGauge(prometheus.GaugeOpts{
		Name: "efs_access_time_ms",
	})

	memoryTotalMbytes.Set(metrics.MemoryTotalMBytes)
	memoryUsedMbytes.Set(float64(metrics.MemoryUsedMBytes))
	memoryUsedPercent.Set(metrics.MemoryUsedPercent)
	cpuLoadPercent.Set(metrics.CPULoadPercent)
	efsAccessTimeMs.Set(float64(metrics.EFSAccessTimeMs))

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

	for _, processMetric := range metrics.Processes {
		processMemoryUsedMbs := promauto.With(registry).NewGauge(prometheus.GaugeOpts{
			Name: "process_memory_used_mbs",
			ConstLabels: map[string]string{
				"name":   processMetric.Name,
				"pid":    strconv.FormatInt(int64(processMetric.PID), 10),
				"parent": strconv.FormatInt(int64(processMetric.ParentPID), 10),
			},
		})

		processMemoryUsedMbs.Set(processMetric.MemoryUsedMbytes)
	}
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

func toMB(value uint64) float64 {
	return float64(value) / 1000 / 1000
}

func getProcessMetrics(ctx context.Context) (processMetrics []ProcessMetric, err error) {
	if processes, err := process.ProcessesWithContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to retrieve the list of processes: %s", err)
	} else {
		for _, p := range processes {
			var name string
			var cmdline string
			var memoryInfo *process.MemoryInfoStat
			var memoryUsed uint64
			var parentPID int32

			name, _ = p.NameWithContext(ctx)
			memoryInfo, _ = p.MemoryInfoWithContext(ctx)
			cmdline, _ = p.CmdlineWithContext(ctx)
			cmdLineSlice, _ := p.CmdlineSliceWithContext(ctx)

			var baseCmdLength float64
			if cmdLineSliceLen := len(cmdLineSlice); cmdLineSliceLen >= 1 {
				baseCmdLength = float64(len(cmdLineSlice[0]))
			}
			max50 := int(math.Min(baseCmdLength+50, float64(len(cmdline))))

			if memoryInfo != nil {
				memoryUsed = memoryInfo.RSS
			}

			parent, _ := p.ParentWithContext(ctx)
			if parent != nil {
				parentPID = parent.Pid
			}

			processMetric := ProcessMetric{
				Name:             name,
				Command:          cmdline[0:max50],
				PID:              p.Pid,
				ParentPID:        parentPID,
				MemoryUsedMbytes: toMB(memoryUsed),
			}

			processMetrics = append(processMetrics, processMetric)
		}
	}

	sort.Slice(processMetrics, func(i, j int) bool {
		return processMetrics[i].MemoryUsedMbytes > processMetrics[j].MemoryUsedMbytes
	})

	top10 := int(math.Min(10, float64(len(processMetrics))))
	return processMetrics[0:top10], nil
}

func CollectMetrics(ctx context.Context, logger log.Logger, startupTimestamp time.Time) (metrics Metrics, aggregateError error) {
	var errors []error

	metrics.BoyarUptime = time.Since(startupTimestamp).Seconds()

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
		metrics.MemoryTotalMBytes = toMB(memInfo.Total)
		metrics.MemoryUsedMBytes = toMB(memInfo.Used)
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

	if processMetrics, processMetricsError := getProcessMetrics(ctx); processMetricsError != nil {
		errors = append(errors, processMetricsError)
	} else {
		metrics.Processes = processMetrics
	}

	return metrics, utils.AggregateErrors(errors)
}
