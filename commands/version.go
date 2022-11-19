package commands

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
	"github.com/symbiosis-cloud/cli/pkg/symcommand"
	"github.com/symbiosis-cloud/symbiosis-go"
)

var (
	Version = ""
)

type VersionCommand struct {
	Client      *symbiosis.Client
	CommandOpts *symcommand.CommandOpts
}

func (n *VersionCommand) Command() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Prints application version",
		Long:  ``,
		Run:   n.Execute,
		PersistentPreRunE: func(command *cobra.Command, args []string) error {
			err := symcommand.Initialise(nodepoolCommands, command)

			if err != nil {
				return err
			}

			return nil
		},
	}
}

func (v *VersionCommand) Execute(command *cobra.Command, args []string) {

	fmt.Printf("Symbiosis CLI version %s", v.GetVersion())
}

func (v *VersionCommand) GetVersion() string {
	if Version == "" {
		buildInfos, ok := debug.ReadBuildInfo()
		if ok && buildInfos.Main.Version != "" {
			return buildInfos.Main.Version
		}
		return "v0.0.0+dev"
	}
	return Version
}

func (v *VersionCommand) Init(client *symbiosis.Client, opts *symcommand.CommandOpts) {
	v.Client = client
	v.CommandOpts = opts
}
