/*
Copyright © 2022 Symbiosis
*/
package cluster

import (
	"github.com/spf13/cobra"
	"github.com/symbiosis-cloud/cli/pkg/symcommand"
	"github.com/symbiosis-cloud/cli/pkg/util"
	"github.com/symbiosis-cloud/symbiosis-go"
)

type ListClusterCommand struct {
	Client      *symbiosis.Client
	CommandOpts *symcommand.CommandOpts
}

func (c *ListClusterCommand) Execute(command *cobra.Command, args []string) error {

	clusters, err := c.Client.Cluster.List(100, 0)

	if err != nil {
		return err
	}

	var data [][]interface{}

	for _, cluster := range clusters.Clusters {
		data = append(data, []interface{}{cluster.ID, cluster.Name, cluster.KubeVersion})
	}

	err = util.NewOutput(util.TableOutput{
		Headers: []string{"ID", "Name", "Version"},
		Data:    data,
	},
		clusters.Clusters,
	).VariableOutput()

	if err != nil {
		return err
	}

	return nil
}

func (c *ListClusterCommand) Command() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List your clusters",
		Long:  ``,
		RunE:  c.Execute,
	}

	return cmd
}

func (c *ListClusterCommand) Init(client *symbiosis.Client, opts *symcommand.CommandOpts) {
	c.Client = client
	c.CommandOpts = opts
}
