/*
Copyright © 2022 Symbiosis Cloud
*/
package commands

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/symbiosis-cloud/cli/pkg/output"
	"github.com/symbiosis-cloud/cli/pkg/symcommand"
	"github.com/symbiosis-cloud/symbiosis-go"
)

var (
	EnableBetaCommands bool
	commands           []symcommand.Command
	cfgFile            string
	verbose            bool
	OutputFormat       string
	yes                bool
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:          "sym <command> <subcommand>",
	Short:        "Easily manage your Symbiosis resources using this CLI application.",
	Long:         ``,
	SilenceUsage: true,
	PersistentPreRunE: func(command *cobra.Command, args []string) error {
		// TODO: find a better way to initialise clients. Because of these commands cannot have pre-runs
		// probably move it to OnInitialize
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

	RootCmd.PersistentFlags().StringP("project", "p", "", "Manually sets the project")
	RootCmd.PersistentFlags().Bool("verbose", false, "Enable verbose logging")
	RootCmd.PersistentFlags().Bool("yes", false, "Skip manual confirmation")
	RootCmd.PersistentFlags().StringVarP(&OutputFormat, "output", "o", "table", "Output format (table, json or yaml). Default: table")

	RootCmd.Flags().StringVar(&cfgFile, "config", "$HOME/.symbiosis/config.yaml", "config file (default is $HOME/.symbiosis/config.yaml)")

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
		&VersionCommand{},
		&TestCommand{},
		&CompletionCommand{},
	}

	// TODO: find a way to toggle beta commands via a flag
	if os.Getenv("SYM_BETA") != "" {
		commands = append(commands, betaCommands...)
		defer text.EnableColors()

		fmt.Printf("%s** NOTE ** You enabled beta commands, these are currently not considered fully functional so use with caution.%s\n", text.FgRed.EscapeSeq(), text.FgWhite.EscapeSeq())
	}

	for _, command := range commands {
		RootCmd.AddCommand(command.Command())
	}

	RootCmd.AddCommand(configCmd)
}

func initConfig() {
	viper.SetConfigType("yaml")

	// set global output format
	output.OutputFormat = OutputFormat

	isDefault := false

	symbiosisApiUrl := symbiosis.APIEndpoint

	if url := os.Getenv("SYMBIOSIS_API_URL"); url != "" {
		symbiosisApiUrl = url
	}

	viper.Set("api_url", symbiosisApiUrl)

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	if cfgFile == "$HOME/.symbiosis/config.yaml" {
		isDefault = true
		cfgFile = home + "/.symbiosis/config.yaml"
	}

	if _, err := os.Stat(cfgFile); os.IsNotExist(err) {

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
			log.Fatalf("Config file %s does not exist, please create it first or use the default.", cfgFile)
		}
	}

	dir, file := path.Split(cfgFile)

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
			log.Fatalf("Config file %s cannot be read", cfgFile)
		} else {
			log.Fatalf("fatal error config file: %v", err)
		}
	}
}
