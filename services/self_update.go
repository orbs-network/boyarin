package services

import (
	"encoding/hex"
	"fmt"
	"github.com/inconshreveable/go-update"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/orbs-network/scribe/log"
	"net/http"
)

// FIXME return actual value
func (coreBoyar *BoyarService) NeedsUpdate() bool {
	return true
}

func (coreBoyar *BoyarService) SelfUpdate(image adapter.ExecutableImageOptions) error {
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

	err = update.Apply(resp.Body, update.Options{
		TargetPath: coreBoyar.binaryTargetPath,
		Checksum:   checksum,
	})
	if err != nil {
		// error handling
	}
	return err
}

func (coreBoyar *BoyarService) CheckForUpdates(flags *config.Flags, cfg config.NodeConfiguration) (shouldExit bool) {
	shouldExit = false
	if flags.AutoUpdate && coreBoyar.NeedsUpdate() {
		if err := coreBoyar.SelfUpdate(cfg.OrchestratorOptions().ExecutableImage); err != nil {
			coreBoyar.logger.Error("failed to update self", log.Error(err))
			return
		}

		return flags.ShutdownAfterUpdate
	}

	return
}
