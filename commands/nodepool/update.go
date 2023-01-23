/*
Copyright © 2022 Symbiosis
*/
package nodepool

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/symbiosis-cloud/cli/pkg/symcommand"
	"github.com/symbiosis-cloud/symbiosis-go"
)

type UpdateNodePoolCommand struct {
	Client      *symbiosis.Client
	CommandOpts *symcommand.CommandOpts
}

var name string

func (c *UpdateNodePoolCommand) Command() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update node-pool",
		Long:  ``,
		PreRunE: func(command *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("Please provide a node pool ID (sym node-pool update <nodePoolId>)")
			}
			return nil
		},
		RunE: func(command *cobra.Command, args []string) error {
			nodePoolId := args[0]
			c.CommandOpts.Logger.Info().Msgf("Updating node-pool %s", nodePoolId)

			var autoscalingSettings symbiosis.AutoscalingSettings

			if autoscaling {
				autoscalingSettings = symbiosis.AutoscalingSettings{
					Enabled: true,
					MinSize: minNodes,
					MaxSize: maxNodes,
				}
				nodeCount = minNodes
			}

			updateInput := &symbiosis.NodePoolUpdateInput{
				Quantity:    nodeCount,
				Autoscaling: autoscalingSettings,
				Name:        name,
			}

			err := c.Client.NodePool.Update(nodePoolId, updateInput)

			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&autoscaling, "autoscaling", false, "Create an autoscaling Node Pool for the cluster")
	cmd.Flags().IntVar(&minNodes, "min-nodes", 2, "Minimum node count. Needs to be at least 2 nodes (used with autoscaling).")
	cmd.Flags().IntVar(&maxNodes, "max-nodes", 4, "Maximum node count. Needs to more than the minimum node count (used with autoscaling).")
	cmd.MarkFlagsRequiredTogether("autoscaling", "min-nodes", "max-nodes")

	cmd.Flags().IntVarP(&nodeCount, "node-count", "n", 1, "Node count for the cluster node pool (ignored when autoscaling is on)")
	cmd.Flags().StringVar(&name, "name", "", "Name of the node pool")

	return cmd
}

func (c *UpdateNodePoolCommand) Init(client *symbiosis.Client, opts *symcommand.CommandOpts) {
	c.Client = client
	c.CommandOpts = opts
}
