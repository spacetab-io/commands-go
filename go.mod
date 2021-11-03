module github.com/spacetab-io/commands-go

go 1.16

require (
	github.com/jackc/pgx/v4 v4.13.0
	github.com/pressly/goose v2.7.0+incompatible
	github.com/spacetab-io/configuration-go v1.2.0
	github.com/spacetab-io/configuration-structs-go v0.0.1
	github.com/spacetab-io/logs-go/v2 v2.1.0
	github.com/spf13/cobra v1.2.1
)

//replace github.com/spacetab-io/configuration-structs-go => ../configuration-structs-go
