/*
Copyright © 2022 Symbiosis
*/
package cluster

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/symbiosis-cloud/cli/pkg/identity"
	"github.com/symbiosis-cloud/cli/pkg/output"
	"github.com/symbiosis-cloud/cli/pkg/symcommand"
	"github.com/symbiosis-cloud/symbiosis-go"
)

type ClusterIdentityCommand struct {
	Client      *symbiosis.Client
	CommandOpts *symcommand.CommandOpts
}

var (
	kubeConfig string
)

func (c *ClusterIdentityCommand) Execute(command *cobra.Command, args []string) error {
	clusterName := args[0]

	identity, err := identity.NewClusterIdentity(c.Client, clusterName, kubeConfig, merge)

	if err != nil {
		return err
	}

	c.CommandOpts.Logger.Info().Msgf("Written identity to %s", identity.KubeConfigPath)

	if kubeConfig == "" {
		fmt.Println(string(identity.Output))
	}

	return nil
}

func (c *ClusterIdentityCommand) Command() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "identity <cluster> [-i path] [--merge]",
		Short: "Retrieve the identity (kubeConfig) for this cluster",
		Long:  ``,
		PreRunE: func(command *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("Please provide a cluster name (sym identity <cluster>)")
			}

			if merge {
				return output.Confirmation("Are you sure you want to merge the new config with your existing .kube/config file", c.CommandOpts.Yes)
			}

			return nil
		},
		RunE: c.Execute,
	}

	cmd.Flags().StringVarP(&kubeConfig, "identity-output-path", "i", "", "Write the generated kubeConfig file to this location")
	cmd.Flags().BoolVar(&merge, "merge", false, "Merge the generated kubeConfig file with the one on your system")
	return cmd
}

func (c *ClusterIdentityCommand) Init(client *symbiosis.Client, opts *symcommand.CommandOpts) {
	c.Client = client
	c.CommandOpts = opts
}
