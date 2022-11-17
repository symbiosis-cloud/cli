/*
Copyright © 2022 Symbiosis
*/
package cluster

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"github.com/symbiosis-cloud/cli/pkg/command"
	"github.com/symbiosis-cloud/symbiosis-go"
	"gopkg.in/yaml.v3"
)

type ListClusterCommand struct {
	Client      *symbiosis.Client
	CommandOpts *command.CommandOpts
}

func (c *ListClusterCommand) Command() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List your clusters",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {

			c, err := c.Client.Cluster.List(100, 0)

			if err != nil {
				return err
			}

			output, err := cmd.Flags().GetString("output")

			if err != nil {
				return err
			}

			if output == "table" {
				t := table.NewWriter()
				t.SetOutputMirror(os.Stdout)
				t.AppendHeader(table.Row{"ID", "Name", "Nodes"})

				if len(c.Clusters) > 0 {
					for _, c := range c.Clusters {
						t.AppendRow([]interface{}{c.ID, c.Name, len(c.Nodes)})
					}
				}

				t.Render()
			} else if output == "json" {
				jsonOutput, err := json.MarshalIndent(c.Clusters, "", "  ")

				if err != nil {
					return err
				}

				fmt.Println(string(jsonOutput))
			} else if output == "yaml" {
				yamlOutput, err := yaml.Marshal(c.Clusters)

				if err != nil {
					return err
				}

				fmt.Println(string(yamlOutput))
			}

			return nil
		},
	}

	return cmd
}

func (c *ListClusterCommand) Init(client *symbiosis.Client, opts *command.CommandOpts) {
	c.Client = client
	c.CommandOpts = opts
}
