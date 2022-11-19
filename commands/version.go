package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/symbiosis-cloud/cli/pkg/symcommand"
	"github.com/symbiosis-cloud/symbiosis-go"
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

func (n *VersionCommand) Execute(command *cobra.Command, args []string) {
	fmt.Printf("Symbiosis CLI version %s", VERSION)
}

func (c *VersionCommand) Init(client *symbiosis.Client, opts *symcommand.CommandOpts) {
	c.Client = client
	c.CommandOpts = opts
}
