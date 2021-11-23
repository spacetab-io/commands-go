package commands

import (
	"errors"
)

type CommandContextKey string

const (
	CommandContextCfgKey        CommandContextKey = "cfg"
	CommandContextCfgKeyDB                        = CommandContextCfgKey + ".db"
	CommandContextCfgKeyLog                       = CommandContextCfgKey + ".log"
	CommandContextCfgKeyAppInfo                   = CommandContextCfgKey + ".appInfo"
	CommandContextCfgKeyStage                     = CommandContextCfgKey + ".stage"
)

const (
	CommandContextObjectKey       CommandContextKey = "obj"
	CommandContextObjectKeySeeder                   = CommandContextObjectKey + ".seeder"
	CommandContextObjectKeyConfig                   = CommandContextObjectKey + ".config"
)

var (
	ErrNoMethodFound   = errors.New("seed method is not exists")
	ErrSeedIsDisabled  = errors.New("seed method is disabled in config")
	ErrBadContextValue = errors.New("context value is empty or has wrong type")
)

// nolint: gochecknoinits // 🤷
func init() {
	MigrateCmd.SetUsageFunc(migrateUsage)
	SeedCmd.SetUsageFunc(seedUsage)

	SeedCmd.AddCommand(SeedRunCmd)
	SeedCmd.AddCommand(SeedRunAllCmd)
	SeedCmd.AddCommand(seedListCmd)
}
