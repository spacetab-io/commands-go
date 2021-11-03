package commands

type CommandContextKey string

const (
	CommandContextCfgKeyOverall CommandContextKey = "cfg"
	CommandContextCfgKeyDB                        = CommandContextCfgKeyOverall + ".db"
	CommandContextCfgKeyLog                       = CommandContextCfgKeyOverall + ".log"
	CommandContextCfgKeyAppInfo                   = CommandContextCfgKeyOverall + ".appInfo"
	CommandContextCfgKeyStage                     = CommandContextCfgKeyOverall + ".stage"
)
