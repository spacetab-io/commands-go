package commands

import (
	"fmt"
	"reflect"

	cfgstructs "github.com/spacetab-io/configuration-structs-go"
)

type Seeder struct {
	seeds   []cfgstructs.SeedInfo
	methods map[string]SeedInterface
}

// SeedMethodType checks SeedInterface implementation and returns Seed object reflect.Type.
func SeedMethodType(seed SeedInterface) reflect.Type {
	return reflect.ValueOf(seed).Elem().Type()
}

// NewSeeder initiates new Seeder object.
func NewSeeder(cfg cfgstructs.SeedingCfg, seeds map[string]reflect.Type, repo interface{}) (SeederInterface, error) {
	s := Seeder{seeds: cfg.Seeds, methods: make(map[string]SeedInterface)}

	for _, seedCfg := range s.seeds {
		if !seedCfg.Enabled {
			continue
		}

		seedType, ok := seeds[seedCfg.ClassName]
		if !ok {
			return nil, fmt.Errorf("%w: %s", ErrSeedClassNameNotRegistered, seedCfg.ClassName)
		}

		seedValue := reflect.New(seedType)
		if !seedValue.IsValid() {
			return nil, fmt.Errorf("%w: %s", ErrSeedClassIsNotValid, seedCfg.ClassName)
		}

		seed, ok := seedValue.Interface().(SeedInterface)
		if !ok || seed == nil {
			return nil, fmt.Errorf("%w: %s", ErrSeedClassNotImplementInterface, seedCfg.ClassName)
		}

		seed.SetCfg(seedCfg)
		seed.SetRepo(repo)

		s.methods[seedCfg.Name] = seed
	}

	return s, nil
}

func (s Seeder) GetMethods() map[string]SeedInterface {
	return s.methods
}

func (s Seeder) GetMethod(name string) (SeedInterface, error) {
	m, ok := s.methods[name]
	if !ok {
		return nil, fmt.Errorf("Seeder.GetMethod %s error: %w", name, ErrNoMethodFound)
	}

	if !m.Enabled() {
		return nil, fmt.Errorf("Seeder.GetMethod %s error: %w", name, ErrSeedIsDisabled)
	}

	return m, nil
}

func (s Seeder) SeedsList() []string {
	result := make([]string, 0)

	for _, seed := range s.seeds {
		status := "+"
		if !seed.Enabled {
			status = "-"
		}

		result = append(result, fmt.Sprintf("[%s] %s - %s", status, seed.Name, seed.Description))
	}

	return result
}
