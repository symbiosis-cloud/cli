/*
Copyright © 2022 Symbiosis
*/
package cmd

import (
	_ "embed"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/symbiosis-cloud/cli/cmd/run"
	"github.com/symbiosis-cloud/cli/pkg/command"
	"github.com/symbiosis-cloud/cli/pkg/identity"
	"github.com/symbiosis-cloud/symbiosis-go"
)

type ApplyCommand struct {
	Client      *symbiosis.Client
	CommandOpts *command.CommandOpts
}

func (c *ApplyCommand) Execute(cmd *cobra.Command, args []string) error {
	kubeConfig, err := cmd.Flags().GetString("identityOutputPath")
	file, err := cmd.Flags().GetString("file")

	if err != nil {
		return err
	}

	runFile, err := run.ReadRunFile(file, c.CommandOpts)

	if err != nil {
		return err
	}

	err = runFile.RunBuilders()

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

	identity, err := identity.NewClusterIdentity(c.Client, clusterName, kubeConfig)

	if err != nil {
		return err
	}

	log.Printf("Written identity to %s\n", identity.KubeConfigPath)
	runFile.SetIdentity(identity)

	err = runFile.RunDeploy()

	if err != nil {
		return err
	}

	log.Println("Apply finished.")

	return nil
}

func (c *ApplyCommand) Command() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Run steps defined in sym.yaml (apply) to an existing cluster",
		Long:  ``,
		RunE:  c.Execute,
	}

	cmd.Flags().String("file", "sym.yaml", "File to use (default: sym.yaml)")
	cmd.Flags().String("identityOutputPath", "", "Write the generated kubeConfig file to this location")

	return cmd
}

func (c *ApplyCommand) Init(client *symbiosis.Client, opts *command.CommandOpts) {
	c.Client = client
	c.CommandOpts = opts
}
