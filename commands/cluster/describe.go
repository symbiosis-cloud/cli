/*
Copyright © 2022 Symbiosis
*/
package cluster

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/symbiosis-cloud/cli/pkg/output"
	"github.com/symbiosis-cloud/cli/pkg/symcommand"
	"github.com/symbiosis-cloud/symbiosis-go"
)

type DescribeClusterCommand struct {
	Client      *symbiosis.Client
	CommandOpts *symcommand.CommandOpts
}

func (c *DescribeClusterCommand) Execute(command *cobra.Command, args []string) error {

	cluster, err := c.Client.Cluster.Describe(args[0])

	if err != nil {
		return err
	}

	err = output.NewOutput(output.TableOutput{
		Headers: []string{"ID", "Name", "Version", "Highly available"},
		Data:    [][]interface{}{{cluster.ID, cluster.Name, cluster.KubeVersion, cluster.IsHighlyAvailable}},
	},
		cluster,
	).VariableOutput()

	if err != nil {
		return err
	}

	return nil
}

func (c *DescribeClusterCommand) Command() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Describe your cluster",
		Long:  ``,
		PreRunE: func(command *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("Please provide a cluster name (sym cluster describe <cluster>")
			}
			return nil
		},
		RunE: c.Execute,
	}

	return cmd
}

func (c *DescribeClusterCommand) Init(client *symbiosis.Client, opts *symcommand.CommandOpts) {
	c.Client = client
	c.CommandOpts = opts
}
