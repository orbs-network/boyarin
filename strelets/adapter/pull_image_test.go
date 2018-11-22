package adapter

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_getAuthForRepositoryReturnsEmptyStringByDefault(t *testing.T) {
	username, password, err := getAuthForRepository("/tmp", "some-image")

	require.EqualError(t, err, "docker credentials not found for image some-image")
	require.Empty(t, username)
	require.Empty(t, password)
}

func Test_getAuthForRepositoryReturnsAuth(t *testing.T) {
	username, password, err := getAuthForRepository("_fixtures", "506367651493.dkr.ecr.us-east-1.amazonaws.com/orbs-network")

	require.NoError(t, err)
	require.Equal(t, "Letov", username)
	require.Equal(t, "RussianFieldOfExperiments\n", password)
}
