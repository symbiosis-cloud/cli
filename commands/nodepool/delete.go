/*
Copyright © 2022 Symbiosis
*/
package nodepool

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/symbiosis-cloud/cli/pkg/output"
	"github.com/symbiosis-cloud/cli/pkg/symcommand"
	"github.com/symbiosis-cloud/symbiosis-go"
)

type DeleteNodePoolCommand struct {
	Client      *symbiosis.Client
	CommandOpts *symcommand.CommandOpts
}

func (c *DeleteNodePoolCommand) Command() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete node-pool",
		Long:  ``,
		PreRunE: func(command *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("Please provide a node pool ID (sym node-pool delete <nodePoolId>)")
			}

			return output.Confirmation(fmt.Sprintf("Are you sure you want want to delete node-pool with ID %s", args[0]))
		},
		RunE: func(command *cobra.Command, args []string) error {
			nodePoolId := args[0]
			c.CommandOpts.Logger.Info().Msgf("Deleting node-pool %s", nodePoolId)

			err := c.Client.NodePool.Delete(nodePoolId)

			if err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}

func (c *DeleteNodePoolCommand) Init(client *symbiosis.Client, opts *symcommand.CommandOpts) {
	c.Client = client
	c.CommandOpts = opts
}
