package config

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestNodeConfigurationContainer_ReloadTimeDelay(t *testing.T) {
	source, err := parseStringConfig("{}")
	require.NoError(t, err)

	source.SetKeyConfigPath("./test/config.json")

	reloadTimeDelay := source.ReloadTimeDelay(15 * time.Minute)
	t.Log(reloadTimeDelay)

	require.Condition(t, func() (success bool) {
		return reloadTimeDelay < 15*time.Minute && reloadTimeDelay > 0
	})
}

func TestNodeConfigurationContainer_ReloadTimeDelayWithNoDelay(t *testing.T) {
	source, err := parseStringConfig("{}")
	require.NoError(t, err)

	source.SetKeyConfigPath("./test/config.json")

	reloadTimeDelay := source.ReloadTimeDelay(0)
	t.Log(reloadTimeDelay)

	require.EqualValues(t, 0, reloadTimeDelay)
}
