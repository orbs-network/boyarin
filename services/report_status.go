package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/orbs-network/boyarin/utils"
	"github.com/orbs-network/boyarin/version"
	"github.com/orbs-network/govnr"
	"github.com/orbs-network/scribe/log"
	"github.com/prometheus/client_golang/prometheus"
	"io/ioutil"
	"time"
)

const SERVICE_STATUS_REPORT_PERIOD = 30 * time.Second
const SERVICE_STATUS_REPORT_TIMEOUT = 15 * time.Second

const MAX_CPU_LOAD = 75
const MAX_MEMORY_LOAD = 75

func WatchAndReportStatusAndMetrics(ctx context.Context, logger log.Logger, flags *config.Flags) govnr.ShutdownWaiter {
	errorHandler := utils.NewLogErrors("service status reporter", logger)
	startupTimestamp := time.Now()
	return govnr.Forever(ctx, "service status reporter", errorHandler, func() {
		start := time.Now()
		ctxWithTimeout, cancel := context.WithTimeout(ctx, SERVICE_STATUS_REPORT_TIMEOUT)
		defer cancel()

		status, metrics := GetStatusAndMetrics(ctxWithTimeout, logger, flags, startupTimestamp, SERVICE_STATUS_REPORT_PERIOD)

		if flags.StatusFilePath != "" {
			rawJSON, _ := json.MarshalIndent(status, "  ", "  ")
			if err := ioutil.WriteFile(flags.StatusFilePath, rawJSON, 0644); err != nil {
				logger.Error("failed to write status file", log.Error(err))
			}
		}

		if flags.MetricsFilePath != "" {
			registry := prometheus.NewRegistry()
			InitializeAndUpdatePrometheusMetrics(registry, metrics)
			if serializedMetrics, err := GetSerializedMetrics(registry); err != nil {
				logger.Error("failed to serialize metrics", log.Error(err))
			} else {
				if err := ioutil.WriteFile(flags.MetricsFilePath, []byte(serializedMetrics), 0644); err != nil {
					logger.Error("failed to write metrics file", log.Error(err))
				}
			}
		}

		logger.Info("finished reporting status")

		select {
		case <-ctx.Done():
		case <-time.After(SERVICE_STATUS_REPORT_PERIOD - time.Since(start)): // to report exactly every minute
		}
	})
}

type StatusResponse struct {
	Timestamp time.Time
	Status    string
	Error     string
	Payload   map[string]interface{}
}

func statusResponseWithError(flags *config.Flags, dockerVersion interface{}, err error) StatusResponse {
	return StatusResponse{
		Status:    "Failed to query Docker Swarm",
		Timestamp: time.Now(),
		Error:     err.Error(),
		Payload: map[string]interface{}{
			"Version": version.GetVersion(),
			"Docker": map[string]interface{}{
				"Version": dockerVersion,
			},
			"Config": flags,
		},
	}
}

func GetStatusAndMetrics(ctx context.Context, logger log.Logger, flags *config.Flags, startupTimestamp time.Time, dockerStatusPeriod time.Duration) (status StatusResponse, metrics Metrics) {
	// We really don't need any options here since we're just observing
	orchestrator, err := adapter.NewDockerSwarm(&adapter.OrchestratorOptions{}, logger)
	if err != nil {
		status = statusResponseWithError(flags, nil, err)
	} else {
		defer orchestrator.Close()

		dockerVersion, _ := orchestrator.Version(ctx)

		containerStatus, err := orchestrator.GetStatus(ctx, dockerStatusPeriod)
		if err != nil {
			status = statusResponseWithError(flags, dockerVersion, err)
		} else {
			services := make(map[string][]*adapter.ContainerStatus)
			for _, s := range containerStatus {
				services[s.Name] = append(services[s.Name], s)
			}

			status = StatusResponse{
				Status:    "OK",
				Timestamp: time.Now(),
				Payload: map[string]interface{}{
					"Version": version.GetVersion(),
					"Docker": map[string]interface{}{
						"Version": dockerVersion,
					},
					"Services": services,
					"Config":   flags,
				},
			}
		}
	}

	metrics, err = CollectMetrics(ctx, logger, startupTimestamp)
	if err != nil {
		status.Status = "Failed to collect metrics"
		status.Error = err.Error()
	}

	if metrics.CPULoadPercent >= 75 {
		status.Status = "CPU usage is too high"
		status.Error = fmt.Sprintf("CPU usage is higher that %d%% (currently at %f%%)",
			MAX_CPU_LOAD, metrics.CPULoadPercent)
	}

	if metrics.MemoryUsedPercent >= 75 {
		status.Status = "Memory usage is too high"
		status.Error = fmt.Sprintf("Memory usage is higher that %d%% (currently at %f%%)",
			MAX_MEMORY_LOAD, metrics.MemoryUsedPercent)
	}

	status.Payload["Metrics"] = metrics

	return
}
