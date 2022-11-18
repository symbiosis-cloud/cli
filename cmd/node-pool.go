package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/symbiosis-cloud/cli/cmd/nodepool"
	"github.com/symbiosis-cloud/cli/pkg/symcommand"
	"github.com/symbiosis-cloud/symbiosis-go"
)

type NodePoolCommand struct {
	Client      *symbiosis.Client
	CommandOpts *symcommand.CommandOpts
}

var nodepoolCommands []symcommand.Command

func (n *NodePoolCommand) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "node-pool <command>",
		Short: "Node pool management",
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

	nodepoolCommands = []symcommand.Command{
		&nodepool.CreateNodePoolCommand{},
		&nodepool.ListNodePoolCommand{},
		&nodepool.DeleteNodePoolCommand{},
		&nodepool.DescribeNodePoolCommand{},
	}

	for _, c := range nodepoolCommands {
		cmd.AddCommand(c.Command())
	}

	return cmd
}

func (n *NodePoolCommand) Execute(command *cobra.Command, args []string) {
	fmt.Println("Available commands: [create, list, delete, identity]")
}

func (c *NodePoolCommand) Init(client *symbiosis.Client, opts *symcommand.CommandOpts) {
	c.Client = client
	c.CommandOpts = opts
}
