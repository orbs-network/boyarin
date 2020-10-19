package services

import (
	"bytes"
	"github.com/c9s/goprocinfo/linux"
	"github.com/orbs-network/boyarin/strelets/adapter"
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
	cpuLoad          prometheus.Gauge
	memoryLoad       prometheus.Gauge
	usedMemoryMBytes prometheus.Gauge
	efsAccessTimeMs  prometheus.Gauge
}

func getCPULoad() (float64, error) {
	cpuStats, err := linux.ReadStat("/proc/stat")
	if err != nil {
		return 0, err
	}
	cpu := cpuStats.CPUStatAll
	return float64(cpu.User + cpu.Nice + cpu.System + cpu.Idle), nil
}

func InitializeMetrics(registry *prometheus.Registry) (Metrics, error) {
	return Metrics{
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
	}, nil
}

const EFS_ACCESS_ERROR_VALUE = float64(-99999)

// in case of error return negative value
func measureEFSAccessTime() (float64, error) {
	start := time.Now()
	err := filepath.Walk(adapter.DEFAULT_EFS_PATH, func(path string, info os.FileInfo, err error) error {
		return err
	})

	if err != nil {
		return EFS_ACCESS_ERROR_VALUE, err
	}

	return float64(time.Since(start).Milliseconds()), nil
}

func CollectMetrics(m Metrics, logger log.Logger) {
	cpu0, errCPU0 := getCPULoad()

	<-time.After(time.Second)

	cpu1, errCPU1 := getCPULoad()
	if errCPU0 != nil {
		logger.Error("failed to read /proc/stat", log.Error(errCPU0))
	} else if errCPU1 != nil {
		logger.Error("failed to read /proc/stat", log.Error(errCPU1))
	} else {
		m.cpuLoad.Set((cpu1 - cpu0) * 100)
	}

	memInfo, err := linux.ReadMemInfo("/proc/meminfo")
	if err != nil {
		logger.Error("failed to read /proc/meminfo", log.Error(err))
	} else {
		m.usedMemoryMBytes.Set(float64(memInfo.MemAvailable-memInfo.MemFree) / 1000)
		m.memoryLoad.Set(float64((memInfo.MemAvailable - memInfo.MemFree) / memInfo.MemAvailable * 100))
	}

	accessTime, err := measureEFSAccessTime()
	if err != nil {
		logger.Error("failed to measure EFS access time", log.Error(err))
	}
	m.efsAccessTimeMs.Set(accessTime)
}
