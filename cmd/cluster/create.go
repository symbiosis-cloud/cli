/*
Copyright © 2022 Symbiosis
*/
package cluster

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/symbiosis-cloud/cli/pkg/identity"
	"github.com/symbiosis-cloud/cli/pkg/symcommand"
	"github.com/symbiosis-cloud/cli/pkg/util"
	"github.com/symbiosis-cloud/symbiosis-go"
)

type CreateClusterCommand struct {
	Client      *symbiosis.Client
	CommandOpts *symcommand.CommandOpts
}

var (
	taints            []string
	labels            []string
	region            string
	version           string
	autoscaling       bool
	isHighlyAvailable bool
	minNodes          int
	maxNodes          int
	nodeType          string
	config            string
	nodeCount         int
	merge             bool
)

func (c *CreateClusterCommand) Command() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "create <cluster name>",
		Short: "Create a cluster",
		Long:  ``,
		PreRunE: func(command *cobra.Command, args []string) error {

			if len(args) == 0 {
				return fmt.Errorf("Please provide a cluster name (sym cluster create <cluster name>)")
			}

			if merge {
				return util.Confirmation("Are you sure you want to merge the new config with your existing .kube/config file")
			}

			return nil
		},
		RunE: func(command *cobra.Command, args []string) error {

			nodeTaints, nodeLabels, err := util.ParseTaintsAndLabels(taints, labels)

			clusterName := args[0]
			c.CommandOpts.Logger.Info().Msgf("Creating cluster %s", clusterName)

			var autoscalingSettings symbiosis.AutoscalingSettings

			if autoscaling {
				autoscalingSettings = symbiosis.AutoscalingSettings{
					Enabled: true,
					MinSize: minNodes,
					MaxSize: maxNodes,
				}
				nodeCount = minNodes
			}

			input := &symbiosis.ClusterInput{
				Name:        clusterName,
				KubeVersion: version,
				Region:      region,
				Nodes: []symbiosis.ClusterNodePoolInput{{
					Name:         fmt.Sprintf("%s-pool-1", clusterName),
					NodeTypeName: nodeType,
					Quantity:     nodeCount,
					Labels:       nodeLabels,
					Taints:       nodeTaints,
					Autoscaling:  autoscalingSettings,
				}},
				IsHighlyAvailable: isHighlyAvailable,
			}

			debugPayload, err := json.MarshalIndent(input, "", " ")

			if err != nil {
				return err
			}

			c.CommandOpts.Logger.Debug().Msgf("Sending payload: %s", string(debugPayload))

			cluster, err := c.Client.Cluster.Create(input)

			if err != nil {
				return err
			}

			c.CommandOpts.Logger.Info().Msgf("Cluster %s created.", cluster.Name)

			if config != "" || merge {
				identity, err := identity.NewClusterIdentity(c.Client, clusterName, config, merge)

				if err != nil {
					return err
				}

				c.CommandOpts.Logger.Info().Msgf("Written identity to %s", identity.KubeConfigPath)

			}

			err = util.NewOutput(util.TableOutput{
				Headers: []string{"ID", "Name", "Version", "Node type", "# Nodes"},
				Data:    [][]interface{}{{cluster.ID, cluster.Name, cluster.KubeVersion, nodeType, nodeCount}},
			},
				cluster,
			).VariableOutput()

			return nil
		},
	}

	cmd.Flags().BoolVar(&isHighlyAvailable, "high-availability", false, "Merge the generated kubeConfig file with the one on your system")

	cmd.Flags().BoolVar(&autoscaling, "autoscaling", false, "Create an autoscaling Node Pool for the cluster")
	cmd.Flags().IntVar(&minNodes, "min-nodes", 2, "Minimum node count. Needs to be at least 2 nodes (used with autoscaling).")
	cmd.Flags().IntVar(&maxNodes, "max-nodes", 4, "Maximum node count. Needs to more than the minimum node count (used with autoscaling).")
	cmd.MarkFlagsRequiredTogether("autoscaling", "min-nodes", "max-nodes")

	cmd.Flags().IntVarP(&nodeCount, "node-count", "n", 1, "Node count for the cluster node pool (ignored when autoscaling is on)")
	cmd.MarkFlagsMutuallyExclusive("autoscaling", "node-count")

	cmd.Flags().StringVarP(&version, "version", "v", "latest", "Kubernetes cluster version")
	cmd.Flags().StringVarP(&region, "region", "r", "germany-1", "Cluster region. Use sym info regions to fetch all regions")
	cmd.Flags().StringVarP(&nodeType, "node-type", "t", "general-1", "Node type. Use sym info node-types to fetch all available node types")

	cmd.Flags().StringArrayVarP(&taints, "taint", "x", []string{}, "Taints to add to the node pool. Example: --taint key=value=NoSchedule")
	cmd.Flags().StringArrayVarP(&labels, "label", "l", []string{}, "Labels to add to the node pool. Example: --label key=value")

	cmd.Flags().BoolVar(&merge, "merge", false, "Merge the generated kubeConfig file with the one on your system.")
	cmd.Flags().StringVarP(&config, "config-path", "k", "", "Kubeconfig output path. By default we do not generate a config file. This can be done manually by running sym cluster identity <cluster>")

	return cmd
}

func (c *CreateClusterCommand) Init(client *symbiosis.Client, opts *symcommand.CommandOpts) {
	c.Client = client
	c.CommandOpts = opts
}
