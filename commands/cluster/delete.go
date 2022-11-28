/*
Copyright © 2022 Symbiosis
*/
package cluster

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/symbiosis-cloud/cli/pkg/output"
	"github.com/symbiosis-cloud/cli/pkg/symcommand"
	"github.com/symbiosis-cloud/symbiosis-go"
)

type DeleteClusterCommand struct {
	Client      *symbiosis.Client
	CommandOpts *symcommand.CommandOpts
}

func (c *DeleteClusterCommand) Command() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete cluster",
		Long:  ``,
		PreRunE: func(command *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("Please provide a cluster name (sym cluster delete <cluster>")
			}

			return output.Confirmation(fmt.Sprintf("Are you sure you want want to delete %s", args[0]))
		},
		RunE: func(command *cobra.Command, args []string) error {
			clusterName := args[0]
			c.CommandOpts.Logger.Info().Msgf("Deleting cluster %s", clusterName)

			err := c.Client.Cluster.Delete(clusterName)

			if err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}

func (c *DeleteClusterCommand) Init(client *symbiosis.Client, opts *symcommand.CommandOpts) {
	c.Client = client
	c.CommandOpts = opts
}
