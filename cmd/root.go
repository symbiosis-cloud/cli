/*
Copyright © 2022 Symbiosis Cloud
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/symbiosis-cloud/cli/pkg/command"
)

var commands []command.Command
var cfgFile string
var debug bool

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "sym <command> <subcommand>",
	Short: "Easily manage your Symbiosis resources using this CLI application.",
	Long:  ``,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		err := command.Initialise(commands, cmd)

		if err != nil {
			return err
		}

		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	RootCmd.PersistentFlags().String("config", "$HOME/.symbiosis/config.yaml", "config file (default is $HOME/.symbiosis/config.yaml)")
	RootCmd.PersistentFlags().StringP("output", "o", "table", "Output format (table, json or yaml). Default: table")
	RootCmd.PersistentFlags().Bool("debug", false, "Enable debug logging")

	commands = []command.Command{
		&ClusterCommand{},
		&RunCommand{},
		&ApplyCommand{},
		&LoginCommand{},
	}

	for _, command := range commands {
		RootCmd.AddCommand(command.Command())
	}

	RootCmd.AddCommand(configCmd)
}
