package commands

import (
	"fmt"

	cfgstructs "github.com/spacetab-io/configuration-structs-go"
	"github.com/spf13/cobra"
)

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	RunE: func(cmd *cobra.Command, args []string) error {
		appInfo, ok := cmd.Context().Value(CommandContextCfgKeyAppInfo).(cfgstructs.ApplicationInfoCfgInterface)
		if !ok {
			return fmt.Errorf("%w: app info config (cfg.appInfo)", ErrBadContextValue)
		}

		cmd.Println(appInfo.GetVersion())

		return nil
	},
}
