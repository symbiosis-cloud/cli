/*
Copyright © 2022 Symbiosis Cloud
*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/symbiosis-cloud/cli/pkg/symcommand"
	"github.com/symbiosis-cloud/cli/pkg/util"
)

var (
	EnableBetaCommands bool
	commands           []symcommand.Command
	cfgFile            string
	debug              bool
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:          "sym <command> <subcommand>",
	Short:        "Easily manage your Symbiosis resources using this CLI application.",
	Long:         ``,
	SilenceUsage: true,
	PersistentPreRunE: func(command *cobra.Command, args []string) error {

		err := symcommand.Initialise(commands, command)

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
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().String("config", "$HOME/.symbiosis/config.yaml", "config file (default is $HOME/.symbiosis/config.yaml)")
	RootCmd.PersistentFlags().StringP("output", "o", "table", "Output format (table, json or yaml). Default: table")
	RootCmd.PersistentFlags().StringP("project", "p", "", "Manually sets the project")
	RootCmd.PersistentFlags().Bool("debug", false, "Enable debug logging")
	RootCmd.PersistentFlags().Bool("yes", false, "Skip manual confirmation")
	// RootCmd.PersistentFlags().Bool("beta", false, "Enable beta features (set to --beta=true to enable)")

	viper.BindPFlag("output", RootCmd.PersistentFlags().Lookup("output"))
	viper.SetDefault("output", string(util.OUTPUT_TABLE))

	viper.BindPFlag("config", RootCmd.PersistentFlags().Lookup("config"))
	viper.BindPFlag("yes", RootCmd.PersistentFlags().Lookup("yes"))

	betaCommands := []symcommand.Command{
		&RunCommand{},
		&ApplyCommand{},
	}

	commands = []symcommand.Command{
		&ClusterCommand{},
		&NodePoolCommand{},
		&LoginCommand{},
		&InfoCommand{},
		&ApiKeysCommand{},
	}

	// TODO: find a way to toggle beta commands via a flag
	if os.Getenv("SYM_BETA") != "" {
		commands = append(commands, betaCommands...)
		fmt.Println("You enabled beta commands, these are currently not considered fully functional so use with caution.")
	}

	for _, command := range commands {
		RootCmd.AddCommand(command.Command())
	}

	RootCmd.AddCommand(configCmd)
}

func initConfig() {
	viper.SetConfigType("yaml")

	isDefault := false

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	configFile := viper.GetString("config")

	if configFile == "$HOME/.symbiosis/config.yaml" {
		isDefault = true
		configFile = home + "/.symbiosis/config.yaml"
	}

	if _, err := os.Stat(configFile); os.IsNotExist(err) {

		if isDefault {

			err = os.Mkdir(home+"/.symbiosis", os.ModePerm)

			if !os.IsExist(err) && err != nil {
				log.Fatal(err)
			}

			viper.AddConfigPath(home + "/.symbiosis")
			viper.SetConfigName("config")

			err = viper.WriteConfigAs(home + "/.symbiosis/config.yaml")

			if err != nil {
				log.Fatal(err)
			}
		} else {
			log.Fatalf("Config file %s does not exist, please create it first or use the default.", configFile)
		}
	}

	dir, file := path.Split(configFile)

	if dir == "" {
		wd, err := os.Getwd()

		if err != nil {
			log.Fatalf(err.Error())
		}

		dir = wd
	}

	configName := strings.TrimSuffix(file, filepath.Ext(file))

	viper.AddConfigPath(dir)
	viper.AddConfigPath(home)
	viper.SetConfigName(configName)

	viper.AutomaticEnv()
	err = viper.ReadInConfig()

	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Fatalf("Config file %s cannot be read", configFile)
		} else {
			log.Fatalf("fatal error config file: %v", err)
		}
	}
}
