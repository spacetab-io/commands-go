package log

import (
	"fmt"
	"os"

	cfgstructs "github.com/spacetab-io/configuration-structs-go/v2"
	"github.com/spacetab-io/configuration-structs-go/v2/contracts"
	log "github.com/spacetab-io/logs-go/v3"
	"go.uber.org/zap"
)

var Logger, _ = log.Init(
	&cfgstructs.Logs{
		Level:   "debug",
		Format:  "text",
		Colored: true,
		Caller:  cfgstructs.CallerConfig{Show: true, SkipFrames: 1},
		Sentry:  nil,
	},
	"unknown",
	"uptimeMaster",
	"unknown",
	os.Stdout,
)

func Init(cfg contracts.LogsCfgInterface, stage, serviceAlias, serviceVersion string) error {
	l, err := log.Init(cfg, stage, serviceAlias, serviceVersion, os.Stdout)
	if err != nil {
		return fmt.Errorf("log init error: %w", err)
	}

	Logger = l

	return nil
}

func Debug() *log.Event { return Logger.Debug() }
func Info() *log.Event  { return Logger.Info() }
func Error() *log.Event { return Logger.Error() }
func GetLogger() *zap.Logger {
	l := *Logger.Logger

	return &l
}
