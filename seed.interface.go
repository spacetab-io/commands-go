package commands

import (
	cfgstructs "github.com/spacetab-io/configuration-structs-go/v2"
)

type SeedInterface interface {
	Enabled() bool
	Name() string
	Seed() error
	SetRepo(r interface{})
	SetCfg(c cfgstructs.SeedInfo)
}

type SeederInterface interface {
	GetMethods() map[string]SeedInterface
	GetMethod(name string) (SeedInterface, error)
	SeedsList() []string
}
