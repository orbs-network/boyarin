package services

import (
	"bytes"
	"github.com/c9s/goprocinfo/linux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/common/expfmt"
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
	cpuLoad          prometheus.Histogram
	memoryLoad       prometheus.Histogram
	usedMemoryMBytes prometheus.Histogram

	diskReadTics  map[string]prometheus.Histogram
	diskWriteTics map[string]prometheus.Histogram
}

func getCPULoad() (float64, error) {
	cpuStats, err := linux.ReadStat("/proc/stat")
	if err != nil {
		return 0, err
	}
	cpu := cpuStats.CPUStatAll
	return float64(cpu.User + cpu.Nice + cpu.System + cpu.Idle), nil
}

func GetDiskNames() (names []string, err error) {
	diskStats, err := readDiskStats()
	for _, stat := range diskStats {
		names = append(names, stat.Name)
	}

	return
}

func InitializeMetrics(registry *prometheus.Registry, diskNames []string) (Metrics, error) {
	usedWriteTics := make(map[string]prometheus.Histogram)
	usedReadTics := make(map[string]prometheus.Histogram)

	for _, name := range diskNames {
		usedWriteTics[name] = promauto.With(registry).NewHistogram(prometheus.HistogramOpts{
			Name: "used_read_ticks",
			Help: "total memory used in megabytes",
			ConstLabels: map[string]string{
				"disk": name,
			},
		})

		usedReadTics[name] = promauto.With(registry).NewHistogram(prometheus.HistogramOpts{
			Name: "used_write_ticks",
			Help: "total memory used in megabytes",
			ConstLabels: map[string]string{
				"disk": name,
			},
		})
	}

	return Metrics{
		usedMemoryMBytes: promauto.With(registry).NewHistogram(prometheus.HistogramOpts{
			Name: "used_memory",
			Help: "total memory used in megabytes",
		}),
		memoryLoad: promauto.With(registry).NewHistogram(prometheus.HistogramOpts{
			Name: "memory_load",
			Help: "total memory load in percent",
		}),
		cpuLoad: promauto.With(registry).NewHistogram(prometheus.HistogramOpts{
			Name: "cpu_load",
			Help: "CPU load across all processors in percent",
		}),
	}, nil
}

func readDiskStats() ([]linux.DiskStat, error) {
	return linux.ReadDiskStats("/proc/diskstats")
}

func findDiskStatByName(stats []linux.DiskStat, name string) (linux.DiskStat, error) {
	for _, stat := range stats {
		if stat.Name == name {
			return stat, nil
		}
	}

	return linux.DiskStat{}, nil
}

const MB = 1000 * 1000
const SECTOR = 512

func collectMetrics(m Metrics) error {
	cpu0, err := getCPULoad()
	if err != nil {
		return nil
	}

	disks0, err := readDiskStats()
	if err != nil {
		return nil
	}

	<-time.After(time.Second)
	cpu1, err := getCPULoad()
	if err != nil {
		return nil
	}

	disks1, err := readDiskStats()
	if err != nil {
		return nil
	}

	m.cpuLoad.Observe((cpu1 - cpu0) * 100)

	memInfo, err := linux.ReadMemInfo("/proc/meminfo")
	if err != nil {
		return nil
	}

	m.usedMemoryMBytes.Observe(float64(memInfo.MemAvailable-memInfo.MemFree) / 1000)
	m.memoryLoad.Observe(float64((memInfo.MemAvailable - memInfo.MemFree) / memInfo.MemAvailable * 100))

	for _, stat0 := range disks0 {
		stat1, err := findDiskStatByName(disks1, stat0.Name)
		if err != nil {
			return err
		}

		diskReadsMbs := float64(stat1.ReadSectors-stat0.ReadSectors) * SECTOR / MB
		diskWritesMbs := float64(stat1.WriteSectors-stat0.WriteSectors) * SECTOR / MB

		m.diskReadTics[stat0.Name].Observe(diskReadsMbs)
		m.diskWriteTics[stat0.Name].Observe(diskWritesMbs)
	}

	return nil
}
