/*
Copyright © 2022 Symbiosis
*/
package apikeys

import (
	"fmt"

	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
	"github.com/symbiosis-cloud/cli/pkg/output"
	"github.com/symbiosis-cloud/cli/pkg/symcommand"
	"github.com/symbiosis-cloud/symbiosis-go"
)

type CreateApiKeyCommand struct {
	Client      *symbiosis.Client
	CommandOpts *symcommand.CommandOpts
}

func (c *CreateApiKeyCommand) Command() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "create <role> <description>",
		Short: "Create an API key",
		Long:  ``,
		PreRunE: func(command *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("Please provide an API key name and description (sym api-key create <name> <description>")
			}

			role := args[0]
			err := symbiosis.ValidateRole(symbiosis.UserRole(role))

			if err != nil {
				return fmt.Errorf("%s is not a valid role. You can run 'sym info roles' to list all available roles.", role)
			}

			return nil
		},
		RunE: func(command *cobra.Command, args []string) error {

			role := args[0]
			description := args[1]

			apiKey, err := c.Client.ApiKeys.Create(symbiosis.ApiKeyInput{
				Description: description,
				Role:        symbiosis.UserRole(role),
			})

			if err != nil {
				return err
			}
			defer text.EnableColors()

			c.CommandOpts.Logger.Info().Msgf("%s** NOTE ** This token will not be shown again.%s", text.FgRed.EscapeSeq(), text.FgWhite.EscapeSeq())

			err = output.NewOutput(output.TableOutput{
				Headers: []string{"ID", "Description", "Token", "Role"},
				Data:    [][]interface{}{{apiKey.ID, apiKey.Description, apiKey.Token, apiKey.Role}},
			},
				apiKey,
			).VariableOutput()

			if err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}

func (c *CreateApiKeyCommand) Init(client *symbiosis.Client, opts *symcommand.CommandOpts) {
	c.Client = client
	c.CommandOpts = opts
}
