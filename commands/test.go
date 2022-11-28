/*
Copyright © 2022 Symbiosis
*/
package commands

import (
	"github.com/spf13/cobra"
	"github.com/symbiosis-cloud/cli/pkg/identity"
	"github.com/symbiosis-cloud/cli/pkg/project"
	"github.com/symbiosis-cloud/cli/pkg/symcommand"
	"github.com/symbiosis-cloud/symbiosis-go"
)

type TestCommand struct {
	Client      *symbiosis.Client
	CommandOpts *symcommand.CommandOpts
}

func (c *TestCommand) Execute(command *cobra.Command, args []string) error {

	_, err := c.Client.Cluster.Describe(args[0])

	if err != nil {
		return err
	}

	clusterName := args[0]

	deploymentFlags, err := symcommand.GetDeploymentFlags(command)

	if err != nil {
		return err
	}

	c.CommandOpts.Namespace = deploymentFlags.Namespace

	identity, err := identity.NewClusterIdentity(c.Client, clusterName, deploymentFlags.IdentityOutputPath, false)

	if err != nil {
		return err
	}

	c.CommandOpts.Logger.Info().Msgf("Written identity to %s", identity.KubeConfigPath)

	projectConfig, err := project.NewProjectConfig(deploymentFlags.File, c.CommandOpts, c.Client, identity)

	if err != nil {
		return err
	}

	err = projectConfig.Parse()

	if err != nil {
		return err
	}

	err = projectConfig.RunTests()

	if err != nil {
		return err
	}

	return nil
}

func (c *TestCommand) Command() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "test <cluster>",
		Short: "Run test steps defined in sym.yaml",
		Long:  ``,
		RunE:  c.Execute,
	}

	symcommand.SetDeploymentFlags(cmd)

	return cmd
}

func (c *TestCommand) Init(client *symbiosis.Client, opts *symcommand.CommandOpts) {
	c.Client = client
	c.CommandOpts = opts
}
