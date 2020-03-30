package config

import (
	"github.com/orbs-network/boyarin/version"
	"github.com/orbs-network/scribe/log"
	"os"
	"time"
)

const DEFAULT_TRUNCATE_WINDOW = 7 * 24 * time.Hour // 1 week

func GetLogger(flags *Flags) (log.Logger, error) {
	var outputs []log.Output

	if flags.LoggerHttpEndpoint != "" {
		outputs = append(outputs, log.NewBulkOutput(
			log.NewHttpWriter(flags.LoggerHttpEndpoint),
			log.NewJsonFormatter().WithTimestampColumn("@timestamp"), 1))
	}

	if flags.LogFilePath != "" {
		logFile, err := os.OpenFile(flags.LogFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}

		fileWriter := log.NewTruncatingFileWriter(logFile, DEFAULT_TRUNCATE_WINDOW)
		outputs = append(outputs, log.NewFormattingOutput(fileWriter, log.NewHumanReadableFormatter()))
	} else {
		outputs = append(outputs, log.NewFormattingOutput(os.Stdout, log.NewHumanReadableFormatter()))
	}

	tags := []*log.Field{
		log.String("app", "boyar"),
		log.String("version", version.GetVersion().Semantic),
		log.String("commit", version.GetVersion().Commit),
	}

	logger := log.GetLogger().
		WithTags(tags...).
		WithOutput(outputs...)

	cfg, _ := NewStringConfigurationSource("{}", "", flags.KeyPairConfigPath, false)
	if err := cfg.VerifyConfig(); err != nil {
		logger.Error("Invalid configuration", log.Error(err))
		return nil, err
	}

	return logger.WithTags(log.Node(string(cfg.NodeAddress()))), nil
}
