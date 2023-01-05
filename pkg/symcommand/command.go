package symcommand

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/symbiosis-cloud/cli/pkg/util"
	"github.com/symbiosis-cloud/cli/pkg/util/firebase"
	"github.com/symbiosis-cloud/symbiosis-go"
	"k8s.io/utils/strings/slices"
)

type Command interface {
	Command() *cobra.Command
	Init(client *symbiosis.Client, opts *CommandOpts)
}

func Initialise(commands []Command, command *cobra.Command) error {

	isAuthCmd := slices.Contains([]string{"init", "login", "version"}, command.CalledAs())
	verbose, err := command.Flags().GetBool("verbose")

	if err != nil {
		return err
	}

	yes, err := command.Flags().GetBool("yes")

	if err != nil {
		return err
	}

	// add commands
	var c *symbiosis.Client

	authMethod := viper.GetString("auth.method")
	refreshToken := viper.GetString("auth.refresh_token")

	if !isAuthCmd {
		err := firebase.ValidateToken(refreshToken)

		if err != nil {
			log.Fatalf(err.Error())
		}
	}

	apiUrl := util.GetEnvOrDefault("SYMBIOSIS_API_URL", symbiosis.APIEndpoint)

	if authMethod == "api_key" {
		client, err := symbiosis.NewClientFromAPIKey(viper.GetString("auth.api_key"), symbiosis.WithEndpoint(apiUrl))

		if err != nil {
			log.Fatalf(err.Error())
		}

		c = client
	} else if authMethod == "token" {
		client, err := symbiosis.NewClientFromToken(viper.GetString("auth.token"), viper.GetString("auth.team_id"), symbiosis.WithEndpoint(apiUrl))

		if err != nil {
			log.Fatalf(err.Error())
		}

		c = client
	}

	if c == nil && !isAuthCmd {
		log.Fatalln("Please authenticate first by running \"sym config init <apiKey>\" or \"sym login\".")
	}

	selectedProject, err := command.Flags().GetString("project")

	if err != nil {
		return err
	}

	var project *symbiosis.Project

	if selectedProject != "" {
		p, err := c.Project.Describe(selectedProject)

		if err != nil {
			return err
		}

		project = p
	}

	output := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}
	output.FormatLevel = func(i interface{}) string {
		return strings.ToUpper(fmt.Sprintf("| %-6s|", i))
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if verbose {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	log := zerolog.New(output).With().Timestamp().Logger()

	opts := &CommandOpts{
		Verbose: verbose,
		Project: project,
		Logger:  log,
		Yes:     yes,
	}

	for _, command := range commands {
		command.Init(c, opts)
	}

	return nil
}
