/*
Copyright © 2022 Symbiosis
*/
package nodepool

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/symbiosis-cloud/cli/pkg/symcommand"
	"github.com/symbiosis-cloud/symbiosis-go"
)

type DescribeNodePoolCommand struct {
	Client      *symbiosis.Client
	CommandOpts *symcommand.CommandOpts
}

func (c *DescribeNodePoolCommand) Command() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "describe",
		Short: "describe node-pool",
		Long:  ``,
		PreRunE: func(command *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("Please provide a cluster name (sym node-pool describe <id>)")
			}

			return nil
		},
		RunE: func(command *cobra.Command, args []string) error {
			nodePoolId := args[0]
			nodePool, err := c.Client.NodePool.Describe(nodePoolId)

			if err != nil {
				return err
			}

			nodePoolOutput([]*symbiosis.NodePool{nodePool}, true)

			return nil
		},
	}

	return cmd
}

func (c *DescribeNodePoolCommand) Init(client *symbiosis.Client, opts *symcommand.CommandOpts) {
	c.Client = client
	c.CommandOpts = opts
}
