package adapter

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_getResourceRequirements(t *testing.T) {
	defaultResourceRequirements := getResourceRequirements(0, 0, 0, 0)
	require.EqualValues(t, 3145728000, defaultResourceRequirements.Limits.MemoryBytes)
	require.EqualValues(t, 0, defaultResourceRequirements.Reservations.MemoryBytes)

	require.EqualValues(t, 1000000000, defaultResourceRequirements.Limits.NanoCPUs)
	require.EqualValues(t, 0, defaultResourceRequirements.Reservations.NanoCPUs)

	limitMemory := getResourceRequirements(100, 0, 0, 0)
	require.EqualValues(t, 100*1024*1024, limitMemory.Limits.MemoryBytes)

	reserveMemory := getResourceRequirements(0, 0, 125, 0)
	require.EqualValues(t, 125*1024*1024, reserveMemory.Reservations.MemoryBytes)

	limitCPU := getResourceRequirements(0, 0.75, 0, 0)
	require.EqualValues(t, int64(0.75*1000000000), limitCPU.Limits.NanoCPUs)

	reserveCPU := getResourceRequirements(0, 0, 0, 2)
	require.EqualValues(t, 2*1000000000, reserveCPU.Reservations.NanoCPUs)
}
