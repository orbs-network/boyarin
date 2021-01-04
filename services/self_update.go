package services

import (
	"encoding/hex"
	"fmt"
	"github.com/inconshreveable/go-update"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/crypto"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/orbs-network/scribe/log"
	"net/http"
)

func (coreBoyar *BoyarService) SelfUpdate(targetPath string, image adapter.ExecutableImageOptions) error {
	checksum, err := hex.DecodeString(image.Sha256)
	if err != nil {
		return fmt.Errorf("could not decode boyar binary SHA256 checksum \"%s\": %s", image.Sha256, err)
	}

	coreBoyar.logger.Info("downloading new boyar binary", log.String("url", image.Url))
	resp, err := http.Get(image.Url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return update.Apply(resp.Body, update.Options{
		TargetPath: targetPath,
		Checksum:   checksum,
	})
}

func (coreBoyar *BoyarService) CheckForUpdates(flags *config.Flags, options adapter.ExecutableImageOptions) (shouldExit bool) {
	shouldExit = false
	if flags.AutoUpdate {
		currentHash, err := crypto.CalculateFileHash(flags.BoyarBinaryPath)
		if err != nil {
			coreBoyar.logger.Error("failed to calculate boyar binary hash", log.Error(err))
			return
		}

		coreBoyar.logger.Info("comparing hashes", log.String("currentHash", currentHash), log.String("updateHash", options.Sha256))
		// always update
		//if currentHash == options.Sha256 { // already the correct version
		//	return
		//}

		if err := coreBoyar.SelfUpdate(flags.BoyarBinaryPath, options); err != nil {
			coreBoyar.logger.Error("failed to update self", log.Error(err))
			return
		} else {
			coreBoyar.logger.Info("successfully replaced boyar binary", log.String("path", flags.BoyarBinaryPath))
		}

		return flags.ShutdownAfterUpdate
	}

	return
}
