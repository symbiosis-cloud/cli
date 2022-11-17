/*
Copyright © 2022 Symbiosis
*/
package cluster

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/symbiosis-cloud/cli/pkg/command"
	"github.com/symbiosis-cloud/symbiosis-go"
)

type DeleteClusterCommand struct {
	Client      *symbiosis.Client
	CommandOpts *command.CommandOpts
}

func (c *DeleteClusterCommand) Command() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete cluster",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("Please provide a cluster name (sym upgrade <cluster>")
			}

			clusterName := args[0]
			log.Printf("Deleting cluster %s", clusterName)

			err := c.Client.Cluster.Delete(clusterName)

			if err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}

func (c *DeleteClusterCommand) Init(client *symbiosis.Client, opts *command.CommandOpts) {
	c.Client = client
	c.CommandOpts = opts
}
