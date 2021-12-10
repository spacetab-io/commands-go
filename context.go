package commands

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
