/*
Copyright © 2022 Symbiosis
*/
package nodepool

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/symbiosis-cloud/cli/pkg/symcommand"
	"github.com/symbiosis-cloud/cli/pkg/util"
	"github.com/symbiosis-cloud/symbiosis-go"
)

type CreateNodePoolCommand struct {
	Client      *symbiosis.Client
	CommandOpts *symcommand.CommandOpts
}

var (
	taints      []string
	labels      []string
	autoscaling bool
	minNodes    int
	maxNodes    int
	nodeType    string
	nodeCount   int
)

func (c *CreateNodePoolCommand) Command() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "create <name> <cluster name>",
		Short: "Create a cluster",
		Long:  ``,
		PreRunE: func(command *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("Please provide a node-pool name and cluster name (sym node-pool create <name> <cluster name>")
			}

			return nil
		},
		RunE: func(command *cobra.Command, args []string) error {

			nodeTaints, nodeLabels, err := util.ParseTaintsAndLabels(taints, labels)

			name := args[0]
			clusterName := args[1]

			c.CommandOpts.Logger.Info().Msgf("Creating node-pool %s in cluster %s", name, clusterName)

			_, err = c.Client.Cluster.Describe(clusterName)

			if err != nil {
				c.CommandOpts.Logger.Debug().Err(err)
				return fmt.Errorf("Failed to retrieve cluster %s. It's possible that the cluster does not exist, please try again.", clusterName)
			}

			var autoscalingSettings symbiosis.AutoscalingSettings

			if autoscaling {
				autoscalingSettings = symbiosis.AutoscalingSettings{
					Enabled: true,
					MinSize: minNodes,
					MaxSize: maxNodes,
				}
				nodeCount = minNodes
			}

			input := &symbiosis.NodePoolInput{
				Name:         name,
				ClusterName:  clusterName,
				NodeTypeName: nodeType,
				Labels:       nodeLabels,
				Taints:       nodeTaints,
				Quantity:     nodeCount,
				Autoscaling:  autoscalingSettings,
			}
			debugPayload, err := json.MarshalIndent(input, "", " ")

			if err != nil {
				return err
			}

			c.CommandOpts.Logger.Debug().Msgf("Sending payload: %s", string(debugPayload))

			nodePool, err := c.Client.NodePool.Create(input)

			if err != nil {
				return err
			}

			c.CommandOpts.Logger.Info().Msgf("Node-pool %s created.", nodePool.Name)

			err = util.NewOutput(util.TableOutput{
				Headers: []string{"ID", "Name", "Type", "Quantity"},
				Data:    [][]interface{}{{nodePool.ID, nodePool.Name, nodeType, nodeCount}},
			},
				nodePool,
			).VariableOutput()

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
	cmd.MarkFlagsMutuallyExclusive("autoscaling", "node-count")

	cmd.Flags().StringVarP(&nodeType, "node-type", "t", "general-1", "Node type. Use sym info node-types to fetch all available node types")

	cmd.Flags().StringArrayVarP(&taints, "taint", "x", []string{}, "Taints to add to the node pool. Example: --taint key=value=NoSchedule")
	cmd.Flags().StringArrayVarP(&labels, "label", "l", []string{}, "Labels to add to the node pool. Example: --label key=value")

	return cmd
}

func (c *CreateNodePoolCommand) Init(client *symbiosis.Client, opts *symcommand.CommandOpts) {
	c.Client = client
	c.CommandOpts = opts
}
