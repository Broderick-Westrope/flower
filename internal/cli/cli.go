package cli

import (
	"errors"

	"github.com/Broderick-Westrope/flower/internal/data"
)

type GlobalDependencies struct {
	Repo *data.Repository
}

type RootCmd struct{}

func (c *RootCmd) Run(deps *GlobalDependencies) error {
	panic(errors.New("unimplemented: TUI"))
}
