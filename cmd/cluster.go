/*
Copyright © 2022 Symbiosis
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/symbiosis-cloud/cli/cmd/cluster"
	"github.com/symbiosis-cloud/cli/pkg/symcommand"
	"github.com/symbiosis-cloud/symbiosis-go"
)

type ClusterCommand struct {
	Client      *symbiosis.Client
	CommandOpts *symcommand.CommandOpts
}

var clusterCommands []symcommand.Command

func (c *ClusterCommand) Command() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Manage your clusters",
		Long:  ``,
		PersistentPreRunE: func(command *cobra.Command, args []string) error {
			err := symcommand.Initialise(clusterCommands, command)

			if err != nil {
				return err
			}

			return nil
		},
		Run: func(command *cobra.Command, args []string) {
			fmt.Println("Available commands: [describe, create, list, delete, identity]")
		},
	}

	clusterCommands = []symcommand.Command{
		&cluster.ListClusterCommand{},
		&cluster.DeleteClusterCommand{},
		&cluster.ClusterIdentityCommand{},
		&cluster.CreateClusterCommand{},
		&cluster.DescribeClusterCommand{},
	}

	for _, c := range clusterCommands {
		cmd.AddCommand(c.Command())
	}

	return cmd
}

func (c *ClusterCommand) Init(client *symbiosis.Client, opts *symcommand.CommandOpts) {
	c.Client = client
	c.CommandOpts = opts
}
