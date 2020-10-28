package services

import (
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/crypto"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/orbs-network/scribe/log"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var VALID_UPDATE_OPTIONS = adapter.ExecutableImageOptions{
	Url:    "https://github.com/orbs-network/boyarin/releases/download/v1.4.0/boyar-v1.4.0.bin",
	Sha256: "1998cc1f7721acfe1954ab2878cc0ad8062cd6d919cd61fa22401c6750e195fe",
}

func TestBoyarService_SelfUpdateHappyFlow(t *testing.T) {
	targetPath, _ := prepareSelfUpdateTest(t)

	logger := log.GetLogger()
	coreBoyar := NewCoreBoyarService(logger)

	flags := &config.Flags{
		AutoUpdate:          true,
		BoyarBinaryPath:     targetPath,
		ShutdownAfterUpdate: true,
	}

	shouldExit := coreBoyar.CheckForUpdates(flags, VALID_UPDATE_OPTIONS)

	require.True(t, shouldExit)
	newChecksum, _ := crypto.CalculateFileHash(targetPath)
	require.EqualValues(t, VALID_UPDATE_OPTIONS.Sha256, newChecksum)
}

func prepareSelfUpdateTest(t *testing.T) (targetPath string, checksum string) {
	os.RemoveAll("./_tmp")
	os.MkdirAll("./_tmp", 0755)

	targetPath = filepath.Join("./_tmp", "boyar.bin")
	err := ioutil.WriteFile(targetPath, []byte("fake binary"), 0755)
	require.NoError(t, err)

	checksum, err = crypto.CalculateFileHash(targetPath)
	require.NoError(t, err)

	return targetPath, checksum
}

func TestBoyarService_SelfUpdateWithDisabledAutoUpdate(t *testing.T) {
	targetPath, checksum := prepareSelfUpdateTest(t)

	logger := log.GetLogger()
	coreBoyar := NewCoreBoyarService(logger)

	flags := &config.Flags{
		AutoUpdate:          false,
		BoyarBinaryPath:     targetPath,
		ShutdownAfterUpdate: true,
	}

	shouldExit := coreBoyar.CheckForUpdates(flags, VALID_UPDATE_OPTIONS)

	// nothing changed
	require.False(t, shouldExit)
	newChecksum, _ := crypto.CalculateFileHash(targetPath)
	require.EqualValues(t, checksum, newChecksum)
}

func TestBoyarService_SelfUpdateWithWrongBinaryPath(t *testing.T) {
	targetPath, checksum := prepareSelfUpdateTest(t)

	logger := log.GetLogger()
	coreBoyar := NewCoreBoyarService(logger)

	flags := &config.Flags{
		AutoUpdate:          true,
		BoyarBinaryPath:     targetPath + "0",
		ShutdownAfterUpdate: true,
	}

	shouldExit := coreBoyar.CheckForUpdates(flags, VALID_UPDATE_OPTIONS)

	// nothing changed
	require.False(t, shouldExit)
	newChecksum, _ := crypto.CalculateFileHash(targetPath)
	require.EqualValues(t, checksum, newChecksum)
}

func TestBoyarService_SelfUpdateWithMalformedUrl(t *testing.T) {
	targetPath, checksum := prepareSelfUpdateTest(t)

	logger := log.GetLogger()
	coreBoyar := NewCoreBoyarService(logger)

	flags := &config.Flags{
		AutoUpdate:          true,
		BoyarBinaryPath:     targetPath,
		ShutdownAfterUpdate: true,
	}

	shouldExit := coreBoyar.CheckForUpdates(flags, adapter.ExecutableImageOptions{
		Url:    "http://localhost/does-not-exist",
		Sha256: VALID_UPDATE_OPTIONS.Sha256,
	})

	// nothing changed
	require.False(t, shouldExit)
	newChecksum, _ := crypto.CalculateFileHash(targetPath)
	require.EqualValues(t, checksum, newChecksum)
}

func TestBoyarService_SelfUpdateWithWrongChecksum(t *testing.T) {
	targetPath, checksum := prepareSelfUpdateTest(t)

	logger := log.GetLogger()
	coreBoyar := NewCoreBoyarService(logger)

	flags := &config.Flags{
		AutoUpdate:          true,
		BoyarBinaryPath:     targetPath,
		ShutdownAfterUpdate: true,
	}

	shouldExit := coreBoyar.CheckForUpdates(flags, adapter.ExecutableImageOptions{
		Url:    VALID_UPDATE_OPTIONS.Url,
		Sha256: "0000cc1f7721acfe1954ab2878cc0ad8062cd6d919cd61fa22401c6750e195fe",
	})

	// nothing changed
	require.False(t, shouldExit)
	newChecksum, _ := crypto.CalculateFileHash(targetPath)
	require.EqualValues(t, checksum, newChecksum)
}
