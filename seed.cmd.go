package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/spacetab-io/commands-go/log"
	"github.com/spf13/cobra"
)

// SeedCmd is a database seeding wrapper command.
var (
	SeedCmd = &cobra.Command{
		Use:       "seed",
		Short:     "Database seeding command",
		ValidArgs: []string{"run", "run-all", "list"},
		Args:      cobra.MinimumNArgs(1),
	}
	seedListCmd = &cobra.Command{
		Use:  "list",
		RunE: seedList,
	}
	SeedRunAllCmd = &cobra.Command{
		Use:  "run-all",
		RunE: seedRunAll,
	}
	SeedRunCmd = &cobra.Command{
		Use:  "run",
		Args: cobra.MinimumNArgs(1),
		RunE: seedRun,
	}
)

// seedUsage shows seed command usage.
// Add it to SeedCmd like SeedCmd.SetUsageFunc(SeedUsage).
func seedUsage(cmd *cobra.Command) error {
	w := cmd.OutOrStderr()
	if _, err := w.Write([]byte(fmt.Sprintf(`Usage:
  %s %s [args]

Args:
  run      runs concreete seed
  run-all  applies all seeds
  list     shows available seeds list
`, cmd.Parent().Name(), cmd.Name()))); err != nil {
		return fmt.Errorf("SeedUsage err: %w", err)
	}

	return nil
}

// seedList returns seeds list.
func seedList(cmd *cobra.Command, _ []string) error {
	s, err := getAppSeeder(cmd.Context())
	if err != nil {
		return err
	}

	cmd.Printf("Available seed list:\n    %s\n", strings.Join(s.SeedsList(), "\n    "))

	return nil
}

// seedRun Execute exact seed for passed method name.
func seedRun(cmd *cobra.Command, args []string) error {
	s, err := getAppSeeder(cmd.Context())
	if err != nil {
		return fmt.Errorf("seedRun getAppSeeder() error: %w", err)
	}

	log.Debug().Strs("seeds", args).Msg("Running seeder...")

	// Execute only the given method names
	for _, item := range args {
		seed, err := s.GetMethod(item)
		if err != nil {
			return fmt.Errorf("seedRun GetMethod error: %w", err)
		}

		if err := seed.Seed(); err != nil {
			return fmt.Errorf("seedRun seed.Seed error: %w", err)
		}
	}

	return nil
}

// seedRunAll Execute all seeds if no method name is given.
func seedRunAll(cmd *cobra.Command, _ []string) error {
	s, err := getAppSeeder(cmd.Context())
	if err != nil {
		return fmt.Errorf("seedRunAll getAppSeeder() error: %w", err)
	}

	log.Debug().Msg("Running all seeder...")

	// We are looping over the method on a Seeder struct
	for _, seed := range s.GetMethods() {
		// Get the method in the current iteration
		// Execute seeder
		if err := seed.Seed(); err != nil {
			return fmt.Errorf("seedRunAll seed.Seed() error: %w", err)
		}
	}

	return nil
}

func getAppSeeder(ctx context.Context) (SeederInterface, error) {
	s, ok := ctx.Value(CommandContextObjectKeySeeder).(SeederInterface)
	if !ok {
		return nil, fmt.Errorf("%w: app seed (cfg.appInfo)", ErrBadContextValue)
	}

	return s, nil
}
