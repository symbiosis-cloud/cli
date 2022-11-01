/*
Copyright © 2022 Symbiosis
*/
package cluster

import (
	"fmt"
	"log"

	_ "embed"

	"github.com/spf13/cobra"
	"github.com/symbiosis-cloud/cli/pkg/command"
	"github.com/symbiosis-cloud/cli/pkg/identity"
	"github.com/symbiosis-cloud/symbiosis-go"
)

type ClusterIdentityCommand struct {
	Client      *symbiosis.Client
	CommandOpts *command.CommandOpts
}

func (c *ClusterIdentityCommand) Command() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "identity",
		Short: "Retrieve the identity (kubeConfig) for this cluster",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("Please provide a cluster name (sym upgrade <cluster>")
			}

			merge, err := cmd.Flags().GetBool("merge")

			if err != nil {
				return err
			}

			if merge {
				// TODO: Implement kubectl merge
				log.Println("Not yet implemented")
			}

			// TODO: add confirmation
			clusterName := args[0]

			kubeConfig, err := cmd.Flags().GetString("identityOutputPath")

			if err != nil {
				return err
			}

			identity, err := identity.NewClusterIdentity(c.Client, clusterName, kubeConfig)

			if err != nil {
				return err
			}

			log.Printf("Written identity to %s", identity.KubeConfigPath)

			return nil
		},
	}

	cmd.Flags().String("identityOutputPath", "", "Write the generated kubeConfig file to this location")
	cmd.Flags().Bool("merge", false, "Merge the generated kubeConfig file with the one on your system")
	return cmd
}

func (c *ClusterIdentityCommand) Init(client *symbiosis.Client, opts *command.CommandOpts) {
	c.Client = client
	c.CommandOpts = opts
}
