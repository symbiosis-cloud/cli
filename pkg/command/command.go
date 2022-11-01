package command

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/symbiosis-cloud/cli/pkg/util/firebase"
	"github.com/symbiosis-cloud/symbiosis-go"
	"k8s.io/utils/strings/slices"
)

type Command interface {
	Command() *cobra.Command
	Init(client *symbiosis.Client, opts *CommandOpts)
}

func Initialise(commands []Command, cmd *cobra.Command) error {
	isAuthCmd := slices.Contains([]string{"init", "login"}, cmd.CalledAs())
	cfgFile, err := cmd.Flags().GetString("config")

	if err != nil {
		return err
	}

	debug, err := cmd.Flags().GetBool("debug")

	if err != nil {
		return err
	}

	err = getConfig(cfgFile)

	if err != nil {
		return err
	}

	// add commands
	var c *symbiosis.Client

	authMethod := viper.GetString("auth.method")

	if !isAuthCmd {
		err := firebase.ValidateToken()

		if err != nil {
			log.Fatalf(err.Error())
		}
	}

	if authMethod == "api_key" {
		client, err := symbiosis.NewClientFromAPIKey(viper.GetString("auth.api_key"))

		if err != nil {
			log.Fatalf(err.Error())
		}

		c = client
	} else if authMethod == "token" {
		client, err := symbiosis.NewClientFromToken(viper.GetString("auth.token"), viper.GetString("auth.team_id"))

		if err != nil {
			log.Fatalf(err.Error())
		}

		c = client
	}

	if c == nil && !isAuthCmd {
		log.Fatalln("Please authenticate first by running \"sym config init <apiKey>\" or \"sym login\".")
	}

	opts := &CommandOpts{
		Debug: debug,
	}

	for _, command := range commands {
		command.Init(c, opts)
	}

	return nil
}

func getConfig(configFile string) error {
	viper.SetConfigType("yaml")

	isDefault := false

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

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
			return fmt.Errorf("Config file %s does not exist, please create it first or use the default.", configFile)
		}
	}

	dir, file := path.Split(configFile)

	if dir == "" {
		wd, err := os.Getwd()

		if err != nil {
			return err
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
			return fmt.Errorf("Config file %s cannot be read", configFile)
		} else {
			log.Fatalf("fatal error config file: %v", err)
		}
	}

	return nil
}
