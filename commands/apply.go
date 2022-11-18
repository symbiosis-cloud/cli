/*
Copyright © 2022 Symbiosis
*/
package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/symbiosis-cloud/cli/pkg/identity"
	"github.com/symbiosis-cloud/cli/pkg/project"
	"github.com/symbiosis-cloud/cli/pkg/symcommand"
	"github.com/symbiosis-cloud/symbiosis-go"
)

type ApplyCommand struct {
	Client      *symbiosis.Client
	CommandOpts *symcommand.CommandOpts
}

func (c *ApplyCommand) Execute(command *cobra.Command, args []string) error {

	deploymentFlags, err := symcommand.GetDeploymentFlags(command)

	if err != nil {
		return err
	}

	c.CommandOpts.Namespace = deploymentFlags.Namespace

	projectConfig, err := project.NewProjectConfig(deploymentFlags.File, c.CommandOpts, c.Client)

	if err != nil {
		return err
	}

	err = projectConfig.Parse()

	if err != nil {
		return err
	}

	err = projectConfig.RunBuilders()

	if err != nil {
		return err
	}

	if len(args) == 0 {
		return fmt.Errorf("Please provide a cluster name (sym apply <cluster>")
	}

	clusterName := args[0]
	_, err = c.Client.Cluster.Describe(clusterName)

	if err != nil {
		return fmt.Errorf("Cluster %s does not exist", clusterName)
	}

	identity, err := identity.NewClusterIdentity(c.Client, clusterName, deploymentFlags.IdentityOutputPath, merge)

	if err != nil {
		return err
	}

	c.CommandOpts.Logger.Info().Msgf("Written identity to %s", identity.KubeConfigPath)

	projectConfig.SetIdentity(identity)

	err = projectConfig.RunDeploy()

	if err != nil {
		return err
	}

	c.CommandOpts.Logger.Info().Msg("Apply finished.")

	return nil
}

func (c *ApplyCommand) Command() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Run steps defined in sym.yaml (apply) to an existing cluster",
		Long:  ``,
		RunE:  c.Execute,
	}

	cmd.Flags().BoolVar(&merge, "merge", false, "Merge the generated kubeConfig file with the one on your system")

	symcommand.SetDeploymentFlags(cmd)

	return cmd
}

func (c *ApplyCommand) Init(client *symbiosis.Client, opts *symcommand.CommandOpts) {
	c.Client = client
	c.CommandOpts = opts
}
