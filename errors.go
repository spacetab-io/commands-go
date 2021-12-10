package commands

import "errors"

var (
	ErrNoMethodFound                  = errors.New("seed method is not exists")
	ErrSeedIsDisabled                 = errors.New("seed method is disabled in config")
	ErrSeedClassNameNotRegistered     = errors.New("seed class name not registered")
	ErrSeedClassIsNotValid            = errors.New("seed class name not valid")
	ErrSeedClassNotImplementInterface = errors.New("seed class name not implements commands.SeedInterface")
	ErrBadContextValue                = errors.New("context value is empty or has wrong type")
)
