/*
Copyright © 2022 Symbiosis
*/
package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/symbiosis-cloud/cli/pkg/symcommand"
	"github.com/symbiosis-cloud/cli/pkg/util"
	"github.com/symbiosis-cloud/symbiosis-go"
	"k8s.io/utils/strings/slices"
)

type InfoCommand struct {
	Client      *symbiosis.Client
	CommandOpts *symcommand.CommandOpts
}

var InfoCommands = []string{"node-types", "regions", "roles"}

func (c *InfoCommand) Execute(command *cobra.Command, args []string) error {

	if len(args) == 0 {
		command.Help()
		return nil
	}

	if ok := slices.Contains(InfoCommands, args[0]); !ok {
		return fmt.Errorf("Unrecognized info command: %s", args[0])
	}

	switch args[0] {
	case "node-types":
		return c.printNodeTypes()
	case "regions":
		return c.printRegions()
	case "roles":
		return c.printRoles()
	}

	return nil
}

func (c *InfoCommand) Command() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "info",
		Short: fmt.Sprintf("Fetch information related to Symbiosis resources. Available commands: %s", strings.Join(InfoCommands, ",")),
		Long:  ``,
		RunE:  c.Execute,
	}

	symcommand.SetDeploymentFlags(cmd)

	return cmd
}

func (c *InfoCommand) Init(client *symbiosis.Client, opts *symcommand.CommandOpts) {
	c.Client = client
	c.CommandOpts = opts
}

func (c *InfoCommand) printNodeTypes() error {

	types, err := c.Client.Node.Types()

	if err != nil {
		return err
	}

	var data [][]interface{}

	for _, nodeType := range types {
		data = append(data, []interface{}{nodeType.ID, nodeType.Name, nodeType.Vcpu, nodeType.MemoryMi})
	}

	err = util.NewOutput(util.TableOutput{
		Headers: []string{"ID", "Name", "vCPU", "Memory"},
		Data:    data,
	},
		types,
	).VariableOutput()

	if err != nil {
		return err
	}

	return nil
}

func (c *InfoCommand) printRegions() error {

	regions, err := c.Client.Region.List()

	if err != nil {
		return err
	}

	var data [][]interface{}

	for _, region := range regions {
		data = append(data, []interface{}{region.ID, region.Name})
	}

	err = util.NewOutput(
		util.TableOutput{
			Headers: []string{"ID", "Name"},
			Data:    data,
		},
		regions,
	).VariableOutput()

	if err != nil {
		return err
	}

	return nil
}

func (c *InfoCommand) printRoles() error {

	roles := symbiosis.GetValidRoles()

	var data [][]interface{}

	for role, valid := range roles {
		data = append(data, []interface{}{role, valid})
	}

	err := util.NewOutput(
		util.TableOutput{
			Headers: []string{"Role", "Valid"},
			Data:    data,
		},
		roles,
	).VariableOutput()

	if err != nil {
		return err
	}

	return nil
}
