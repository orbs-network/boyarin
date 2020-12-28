package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/test/helpers"
	"github.com/orbs-network/scribe/log"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestDockerVersion(t *testing.T) {
	//helpers.SkipUnlessSwarmIsEnabled(t)

	helpers.WithContext(func(ctx context.Context) {
		logger := log.DefaultTestingLogger(t)

		status, _ := GetStatusAndMetrics(ctx, logger, &config.Flags{
			ConfigUrl: "http://some/fake/url",
		}, time.Now(), 5*time.Second)

		require.Regexp(t, "RAM.*CPU.*EFSAccess.*", status.Status)

		version := status.Payload["Docker"]
		require.NotNil(t, version)

		raw, _ := json.MarshalIndent(status, "", "  ")
		fmt.Println(string(raw))
	})
}
