package config

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestNodeConfigurationContainer_ReloadTimeDelay(t *testing.T) {
	source, err := NewStringConfigurationSource("{}", "", fakeKeyPair)
	require.NoError(t, err)

	reloadTimeDelay := source.ReloadTimeDelay(15 * time.Minute)
	t.Log(reloadTimeDelay)

	require.Condition(t, func() (success bool) {
		return reloadTimeDelay < 15*time.Minute && reloadTimeDelay > 0
	})
}

func TestNodeConfigurationContainer_ReloadTimeDelayWithNoDelay(t *testing.T) {
	source, err := NewStringConfigurationSource("{}", "", fakeKeyPair)
	require.NoError(t, err)

	reloadTimeDelay := source.ReloadTimeDelay(0)
	t.Log(reloadTimeDelay)

	require.EqualValues(t, 0, reloadTimeDelay)
}

func TestNodeConfigurationContainer_ReloadTimeDelayWithOverride(t *testing.T) {
	source, err := NewStringConfigurationSource(`{"orchestrator": {"max-reload-time-delay": "1m"}}`, "", fakeKeyPair)
	require.NoError(t, err)

	reloadTimeDelay := source.ReloadTimeDelay(0)
	t.Log(reloadTimeDelay)

	require.EqualValues(t, 1*time.Minute, reloadTimeDelay)
}
