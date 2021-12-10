package commands

// nolint: gochecknoinits // 🤷
func init() {
	MigrateCmd.SetUsageFunc(migrateUsage)
	SeedCmd.SetUsageFunc(seedUsage)

	SeedCmd.AddCommand(SeedRunCmd)
	SeedCmd.AddCommand(SeedRunAllCmd)
	SeedCmd.AddCommand(seedListCmd)
}
