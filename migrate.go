package commands

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/log/zerologadapter"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/pressly/goose"
	cfgstructs "github.com/spacetab-io/configuration-structs-go"
	log "github.com/spacetab-io/logs-go/v2"
	"github.com/spf13/cobra"
)

const (
	failureCode = 1
)

var (
	// migrateCmd is a github.com/pressly/goose database migrate wrapper command
	// nolint:deadcode,unused,varcheck
	migrateCmd = &cobra.Command{
		Use:       "migrate",
		Short:     "Database migrations command",
		ValidArgs: []string{"up", "up-by-one", "up-to", "create", "down", "down-to", "fix", "redo", "reset", "status", "version"},
		Args:      cobra.MinimumNArgs(1),
		RunE:      migrate,
	}
	ErrBadContextValue = errors.New("context value is empty or has wrong type")
)

// migrateUsage shows command usage.
// Add it to migrateCmd like migrateCmd.SetUsageFunc(migrateUsage).
// nolint:deadcode,unused
func migrateUsage(cmd *cobra.Command) error {
	w := cmd.OutOrStderr()
	_, err := w.Write([]byte(`Usage:
  <service> migrate [args]

Args:
  up     migrate all upwards
  down   migrate last migration down
`))

	return fmt.Errorf("migrateUsage err: %w", err)
}

const errStrFormat = "%s %s error: %w"

// migrate is a function for cobra.Command RunE param.
func migrate(cmd *cobra.Command, args []string) error {
	method := "migrate"

	envStage, dbCfg, logCfg, appInfo, err := getConfigs(cmd.Context())
	if err != nil {
		return fmt.Errorf(errStrFormat, method, "getConfig", err)
	}

	if err := log.Init(envStage, logCfg, appInfo.GetAlias(), appInfo.GetVersion(), os.Stdout); err != nil {
		log.Error().Err(err).Send()

		return fmt.Errorf(errStrFormat, method, "log.Init", err)
	}

	command := args[0]

	log.Debug().Str("command", command).Strs("command args", args[0:]).Msg("run migrate command")

	cfg, err := pgx.ParseConfig(dbCfg.GetDSN())
	if err != nil {
		log.Error().Err(err).Str("dsn", dbCfg.GetMigrationDSN()).Msg("fail to parse config")

		return fmt.Errorf(errStrFormat, method, "ParseConfig", err)
	}

	cfg.Logger = zerologadapter.NewLogger(log.Logger().With().CallerWithSkipFrameCount(4).Logger()) // nolint:gomnd
	cfg.PreferSimpleProtocol = true

	stdlib.RegisterConnConfig(cfg)

	db := stdlib.OpenDB(*cfg)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("migrate SetDialect error: %w", err)
	}

	if err := db.Ping(); err != nil {
		log.Error().Err(err).Str("dsn", dbCfg.GetDSN()).Msg("fail to ping database")

		return fmt.Errorf("migrate Ping error: %w", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Error().Str("dsn", dbCfg.GetDSN()).Err(err).Msg("fail to close DB connection")

			os.Exit(failureCode)
		}
	}()

	var arguments []string

	// nolint:gomnd
	if len(args) > 3 {
		arguments = append(arguments, args[3:]...)
	}

	if err := checkInit(dbCfg, db); err != nil {
		log.Error().Str("dsn", dbCfg.GetDSN()).Err(err).Msg("fail to check db")

		return fmt.Errorf("migrate checkInit error: %w", err)
	}

	return goose.Run(command, db, dbCfg.GetMigrationsPath(), arguments...)
}

func checkInit(cfg cfgstructs.DatabaseCfgInterface, db *sql.DB) error {
	log.Trace().Msg("detect goose table exists")

	var t *string

	r := db.QueryRow(fmt.Sprintf("SELECT to_regclass('%s.%s') as t", cfg.GetSchema(), cfg.GetMigrationsTableName()))
	if err := r.Scan(&t); err != nil {
		log.Error().Err(err).Str("dsn", cfg.GetDSN()).Msgf("fail to check migrations table")

		return fmt.Errorf("checkInit db.QueryRow error: %w", err)
	}

	//nolint: lll
	create := fmt.Sprintf(
		`
CREATE SEQUENCE IF NOT EXISTS %s_id_seq;
CREATE TABLE %s.%s ("id" int4 NOT NULL DEFAULT nextval('goose_db_version_id_seq'::regclass),"version_id" int8 NOT NULL,"is_applied" bool NOT NULL,"tstamp" timestamp DEFAULT now(),PRIMARY KEY ("id"));
INSERT INTO %s.%s ("version_id", "is_applied", "tstamp") VALUES ('0', 't', NOW());`,
		cfg.GetMigrationsTableName(),
		cfg.GetSchema(),
		cfg.GetMigrationsTableName(),
		cfg.GetSchema(),
		cfg.GetMigrationsTableName(),
	)

	if t == nil {
		log.Trace().Msg("goose table doesn't exists. let's create it")

		_, err := db.Exec(create)
		if err != nil {
			return fmt.Errorf("checkInit db.Exec error: %w", err)
		}

		log.Trace().Msg("goose table now exists. continue")

		return nil
	}

	log.Trace().Msg("goose table exists. go forward")

	return nil
}

func getConfigs(ctx context.Context) (
	string,
	cfgstructs.DatabaseCfgInterface,
	log.Config,
	cfgstructs.ApplicationInfoCfgInterface,
	error,
) {
	envStage, ok := ctx.Value("cfg.envStage").(string)
	if !ok {
		return "", nil, log.Config{}, nil, fmt.Errorf("%w: stage name (cfg.envStage)", ErrBadContextValue)
	}

	dbCfg, ok := ctx.Value("cfg.db").(cfgstructs.DatabaseCfgInterface)
	if !ok {
		return "", nil, log.Config{}, nil, fmt.Errorf("%w: database config (cfg.db)", ErrBadContextValue)
	}

	logCfg, ok := ctx.Value("cfg.log").(log.Config)
	if !ok {
		return "", nil, log.Config{}, nil, fmt.Errorf("%w: log config (cfg.log)", ErrBadContextValue)
	}

	appInfo, ok := ctx.Value("cfg.appInfo").(cfgstructs.ApplicationInfoCfgInterface)
	if !ok {
		return "", nil, log.Config{}, nil, fmt.Errorf("%w: app info config (cfg.appInfo)", ErrBadContextValue)
	}

	return envStage, dbCfg, logCfg, appInfo, nil
}
