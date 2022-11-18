package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/symbiosis-cloud/cli/commands/apikeys"
	"github.com/symbiosis-cloud/cli/pkg/symcommand"
	"github.com/symbiosis-cloud/symbiosis-go"
)

type ApiKeysCommand struct {
	Client      *symbiosis.Client
	CommandOpts *symcommand.CommandOpts
}

var apiKeysCommands []symcommand.Command

func (n *ApiKeysCommand) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api-keys <command>",
		Short: "API key management",
		Long:  ``,
		Run:   n.Execute,
		PersistentPreRunE: func(command *cobra.Command, args []string) error {
			err := symcommand.Initialise(apiKeysCommands, command)

			if err != nil {
				return err
			}

			return nil
		},
	}

	apiKeysCommands = []symcommand.Command{
		&apikeys.CreateApiKeyCommand{},
		&apikeys.DeleteApiKeyCommand{},
		&apikeys.ListApiKeysCommand{},
	}

	for _, c := range apiKeysCommands {
		cmd.AddCommand(c.Command())
	}

	return cmd
}

func (n *ApiKeysCommand) Execute(command *cobra.Command, args []string) {
	fmt.Println("Available commands: [create, list, delete]")
}

func (c *ApiKeysCommand) Init(client *symbiosis.Client, opts *symcommand.CommandOpts) {
	c.Client = client
	c.CommandOpts = opts
}
