/*
Copyright © 2022 Symbiosis
*/
package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/symbiosis-cloud/cli/commands/config"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "General configuration for the CLI tool",
	Long:  ``,
	Run: func(command *cobra.Command, args []string) {
		fmt.Println("Available commands: [init]")
	},
}

func init() {
	configCmd.AddCommand(config.Init)
}
