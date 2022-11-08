package commands

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	zapadapter "github.com/jackc/pgx-zap"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jackc/pgx/v5/tracelog"
	"github.com/pressly/goose"
	"github.com/spacetab-io/commands-go/log"
	"github.com/spacetab-io/configuration-go/stage"
	"github.com/spacetab-io/configuration-structs-go/v2/contracts"
	log2 "github.com/spacetab-io/logs-go/v3"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

const (
	CmdFailureCode  = 1
	CmdErrStrFormat = "%s %s error: %w"
)

// MigrateCmd is a github.com/pressly/goose database migrate wrapper command.
var MigrateCmd = &cobra.Command{
	Use:       "migrate",
	Short:     "Database migrations command",
	ValidArgs: []string{"up", "up-by-one", "up-to", "create", "down", "down-to", "fix", "redo", "reset", "status", "version"},
	Args:      cobra.MinimumNArgs(1),
	RunE:      migrate,
}

// migrateUsage shows command usage.
func migrateUsage(cmd *cobra.Command) error {
	w := cmd.OutOrStderr()
	if _, err := w.Write([]byte(fmt.Sprintf(`Usage:
  %s %s [args]

Args:
  create      writes a new blank migration file
  up          applies all available migrations
  up-by-one   migrates up by a single version
  up-to       migrates up to a specific version
  down        rolls back a single migration from the current version
  down-to     rolls back migrations to a specific version
  fix         fixes migrations file name (?)
  redo        rolls back the most recently applied migration, then runs it again
  reset       rolls back all migrations
  status      prints the status of all migrations
  version     prints the current version of the database
`, cmd.Parent().Name(), cmd.Name()))); err != nil {
		return fmt.Errorf("migrateUsage err: %w", err)
	}

	return nil
}

// migrate is a function for cobra.Command RunE param.
func migrate(cmd *cobra.Command, args []string) error {
	method := "migrate"

	appStage, dbCfg, logCfg, appInfo, err := getConfigs(cmd.Context())
	if err != nil {
		return fmt.Errorf(CmdErrStrFormat, method, "getConfig", err)
	}

	if err := log.Init(logCfg, appStage.String(), appInfo.GetAlias(), appInfo.GetVersion()); err != nil {
		log.Error().Err(err).Send()

		return fmt.Errorf(CmdErrStrFormat, method, "log.Init", err)
	}

	log.Info().Msg(appInfo.Summary())

	command := args[0]

	log.Debug().Str("command", command).Strs("command args", args[0:]).Msg("run migrate command")

	pgxConfig, err := pgx.ParseConfig(dbCfg.GetDSN())
	if err != nil {
		log.Error().Err(err).Str("dsn", dbCfg.GetMigrationDSN()).Msg("fail to parse config")

		return fmt.Errorf(CmdErrStrFormat, method, "ParseConfig", err)
	}

	trLevel, err := tracelog.LogLevelFromString(log.Logger.Level.String())
	if err != nil {
		return fmt.Errorf("log level from string error: %w", err)
	}

	pgxConfig.Tracer = &tracelog.TraceLog{
		Logger:   zapadapter.NewLogger(log.GetLogger().WithOptions(zap.WithCaller(false))),
		LogLevel: trLevel,
	}
	pgxConfig.RuntimeParams = map[string]string{"standard_conforming_strings": "on"}

	stdlib.RegisterConnConfig(pgxConfig)

	db := stdlib.OpenDB(*pgxConfig)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("migrate SetDialect error: %w", err)
	}

	// set migrations table from cfg
	goose.SetTableName(dbCfg.GetMigrationsTableName())

	goose.SetLogger(log2.NewGooseLogger(log.Logger))

	if err := db.Ping(); err != nil {
		log.Error().Err(err).Str("dsn", dbCfg.GetDSN()).Msg("fail to ping database")

		return fmt.Errorf("migrate Ping error: %w", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Error().Str("dsn", dbCfg.GetDSN()).Err(err).Msg("fail to close DB connection")

			os.Exit(CmdFailureCode)
		}
	}()

	var arguments []string

	//nolint:gomnd
	if len(args) > 3 {
		arguments = append(arguments, args[3:]...)
	}

	if err := checkInit(dbCfg, db); err != nil {
		log.Error().Str("dsn", dbCfg.GetDSN()).Err(err).Msg("fail to check db")

		return fmt.Errorf("migrate checkInit error: %w", err)
	}

	if err := goose.Run(command, db, dbCfg.GetMigrationsPath(), arguments...); err != nil {
		return fmt.Errorf("goose run error: %w", err)
	}

	return nil
}

func checkInit(cfg contracts.DatabaseCfgInterface, db *sql.DB) error {
	log.Debug().Msg("detect goose table exists")

	var t *string

	r := db.QueryRow(fmt.Sprintf("SELECT to_regclass('%s.%s') as t", cfg.GetSchema(), cfg.GetMigrationsTableName()))
	if err := r.Scan(&t); err != nil {
		log.Error().Err(err).Str("dsn", cfg.GetDSN()).Msgf("fail to check migrations table")

		return fmt.Errorf("checkInit db.QueryRow error: %w", err)
	}

	//nolint:lll,gosec // lll – for pretty view, gosec – for simplicity in query building
	create := fmt.Sprintf(
		`
CREATE SEQUENCE IF NOT EXISTS %s_id_seq;
CREATE TABLE %s.%s ("id" int4 NOT NULL DEFAULT nextval('%s_id_seq'::regclass),"version_id" int8 NOT NULL,"is_applied" bool NOT NULL,"tstamp" timestamp DEFAULT now(),PRIMARY KEY ("id"));
INSERT INTO %s.%s ("version_id", "is_applied", "tstamp") VALUES ('0', 't', NOW());`,
		cfg.GetMigrationsTableName(),
		cfg.GetSchema(),
		cfg.GetMigrationsTableName(),
		cfg.GetMigrationsTableName(),
		cfg.GetSchema(),
		cfg.GetMigrationsTableName(),
	)

	if t == nil {
		log.Debug().Msg("goose table doesn't exists. let's create it")

		if _, err := db.Exec(create); err != nil {
			return fmt.Errorf("checkInit db.Exec error: %w", err)
		}

		log.Debug().Msg("goose table now exists. continue")

		return nil
	}

	log.Debug().Msg("goose table exists. go forward")

	return nil
}

//nolint:ireturn // it is proxy-method than returns interfaces
func getConfigs(ctx context.Context) (
	stage.Interface,
	contracts.DatabaseCfgInterface,
	contracts.LogsCfgInterface,
	contracts.ApplicationInfoCfgInterface,
	error,
) {
	appStage, ok := ctx.Value(CommandContextCfgKeyStage).(stage.Interface)
	if !ok {
		return nil, nil, nil, nil, fmt.Errorf("%w: stage name (cfg.envStage)", ErrBadContextValue)
	}

	logCfg, ok := ctx.Value(CommandContextCfgKeyLog).(contracts.LogsCfgInterface)
	if !ok {
		return nil, nil, nil, nil, fmt.Errorf("%w: log config (cfg.log)", ErrBadContextValue)
	}

	appInfoCfg, ok := ctx.Value(CommandContextCfgKeyAppInfo).(contracts.ApplicationInfoCfgInterface)
	if !ok {
		return nil, nil, nil, nil, fmt.Errorf("%w: app info config (cfg.appInfo)", ErrBadContextValue)
	}

	dbCfg, ok := ctx.Value(CommandContextCfgKeyDB).(contracts.DatabaseCfgInterface)
	if !ok {
		return nil, nil, nil, nil, fmt.Errorf("%w: database config (cfg.db)", ErrBadContextValue)
	}

	return appStage, dbCfg, logCfg, appInfoCfg, nil
}
