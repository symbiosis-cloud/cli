/*
Copyright © 2022 Symbiosis
*/
package apikeys

import (
	"github.com/spf13/cobra"
	"github.com/symbiosis-cloud/cli/pkg/output"
	"github.com/symbiosis-cloud/cli/pkg/symcommand"
	"github.com/symbiosis-cloud/symbiosis-go"
)

type ListApiKeysCommand struct {
	Client      *symbiosis.Client
	CommandOpts *symcommand.CommandOpts
}

func (c *ListApiKeysCommand) Execute(command *cobra.Command, args []string) error {

	apiKeys, err := c.Client.ApiKeys.List()

	if err != nil {
		return err
	}

	var data [][]interface{}

	for _, apiKey := range apiKeys {
		data = append(data, []interface{}{apiKey.ID, apiKey.Description, apiKey.Token, apiKey.Role})
	}

	err = output.NewOutput(output.TableOutput{
		Headers: []string{"ID", "Description", "Token", "Role"},
		Data:    data,
	},
		apiKeys,
	).VariableOutput()

	if err != nil {
		return err
	}

	return nil
}

func (c *ListApiKeysCommand) Command() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List your api-keys",
		Long:  ``,
		RunE:  c.Execute,
	}

	return cmd
}

func (c *ListApiKeysCommand) Init(client *symbiosis.Client, opts *symcommand.CommandOpts) {
	c.Client = client
	c.CommandOpts = opts
}
