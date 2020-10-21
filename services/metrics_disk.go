// +build !darwin

package services

import (
	"fmt"
	"github.com/orbs-network/scribe/log"
	"github.com/shirou/gopsutil/disk"
	"strings"
)

func ReadDiskMetrics(logger log.Logger) (diskMetrics []DiskMetric, errors []error) {
	if partitions, err := disk.Partitions(false); err != nil {
		errors = append(errors, fmt.Errorf("failed to read partitions info: %s", err))
		logger.Error("failed to read partitions info", log.Error(err))
	} else {
		for _, partition := range partitions {
			if strings.HasPrefix(partition.Mountpoint, "/snap") {
				continue
			}

			if usage, err := disk.Usage(partition.Mountpoint); err != nil {
				errors = append(errors, fmt.Errorf("failed to read disk usage info: %s", err))
				logger.Error("failed to read partitions info", log.Error(err))
			} else {
				diskMetrics = append(diskMetrics, DiskMetric{
					Mountpoint:  partition.Mountpoint,
					UsedPercent: usage.UsedPercent,
					UsedMbytes:  float64(usage.Used) / 1000 / 1000,
					TotalMbytes: float64(usage.Total) / 1000 / 1000,
				})
			}
		}
	}

	return
}
