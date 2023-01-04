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

type ListNodePoolCommand struct {
	Client      *symbiosis.Client
	CommandOpts *symcommand.CommandOpts
}

func (c *ListNodePoolCommand) Command() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List node-pools",
		Long:  ``,
		PreRunE: func(command *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("Please provide a cluster name (sym node-pool list <cluster>)")
			}

			return nil
		},
		RunE: func(command *cobra.Command, args []string) error {
			clusterName := args[0]
			cluster, err := c.Client.Cluster.Describe(clusterName)

			if err != nil {
				return err
			}

			c.CommandOpts.Logger.Info().Msgf("Listing %d node pools for cluster %s", len(cluster.NodePools), cluster.Name)
			nodePoolOutput(cluster.NodePools, false)

			return nil
		},
	}

	return cmd
}

func nodePoolOutput(nodepools []*symbiosis.NodePool, isSingular bool) error {
	var data [][]interface{}

	for _, nodePool := range nodepools {

		var row []interface{}

		row = append(row, nodePool.ID, nodePool.Name, nodePool.NodeTypeName, nodePool.DesiredQuantity)

		if nodePool.Autoscaling.Enabled {
			row = append(row, "On", nodePool.Autoscaling.MinSize, nodePool.Autoscaling.MaxSize)
		} else {
			row = append(row, "Off", "", "")
		}
		data = append(data, row)
	}

	var dataOutput interface{}
	dataOutput = nodepools

	if isSingular {
		dataOutput = nodepools[0]
	}

	err := output.NewOutput(output.TableOutput{
		Headers: []string{"ID", "Name", "Type", "# Nodes", "Autoscaling", "Min nodes", "Nax nodes"},
		Data:    data,
	},
		dataOutput,
	).VariableOutput()

	return err
}

func (c *ListNodePoolCommand) Init(client *symbiosis.Client, opts *symcommand.CommandOpts) {
	c.Client = client
	c.CommandOpts = opts
}
