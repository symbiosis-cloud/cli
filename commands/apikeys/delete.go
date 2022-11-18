/*
Copyright © 2022 Symbiosis
*/
package apikeys

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/symbiosis-cloud/cli/pkg/symcommand"
	"github.com/symbiosis-cloud/cli/pkg/util"
	"github.com/symbiosis-cloud/symbiosis-go"
)

type DeleteApiKeyCommand struct {
	Client      *symbiosis.Client
	CommandOpts *symcommand.CommandOpts
}

func (c *DeleteApiKeyCommand) Command() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete api-key",
		Long:  ``,
		PreRunE: func(command *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("Please provide an api-key ID (sym api-key delete <id>")
			}

			return util.Confirmation(fmt.Sprintf("Are you sure you want want to delete api-key %s", args[0]))
		},
		RunE: func(command *cobra.Command, args []string) error {
			apiKeyId := args[0]
			c.CommandOpts.Logger.Info().Msgf("Deleting api-key %s", apiKeyId)

			err := c.Client.ApiKeys.Delete(apiKeyId)

			if err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}

func (c *DeleteApiKeyCommand) Init(client *symbiosis.Client, opts *symcommand.CommandOpts) {
	c.Client = client
	c.CommandOpts = opts
}
