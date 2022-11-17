package symcommand

import (
	"github.com/rs/zerolog"
	"github.com/symbiosis-cloud/symbiosis-go"
)

type CommandOpts struct {
	Debug     bool
	Namespace string
	Project   *symbiosis.Project
	Logger    zerolog.Logger
}
