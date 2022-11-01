/*
Copyright © 2022 Symbiosis
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/symbiosis-cloud/cli/cmd/cluster"
	"github.com/symbiosis-cloud/cli/pkg/command"
	"github.com/symbiosis-cloud/symbiosis-go"
)

type ClusterCommand struct {
	Client      *symbiosis.Client
	CommandOpts *command.CommandOpts
}

var clusterCommands []command.Command

func (c *ClusterCommand) Command() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Manage your clusters",
		Long:  ``,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			err := command.Initialise(clusterCommands, cmd)

			if err != nil {
				return err
			}

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Available commands: [create, list, delete, identity]")
		},
	}

	clusterCommands = []command.Command{
		&cluster.ListClusterCommand{},
		&cluster.DeleteClusterCommand{},
		&cluster.ClusterIdentityCommand{},
	}

	for _, command := range clusterCommands {
		cmd.AddCommand(command.Command())
	}

	return cmd
}

func (c *ClusterCommand) Init(client *symbiosis.Client, opts *command.CommandOpts) {
	c.Client = client
	c.CommandOpts = opts
}
