/*
Copyright © 2022 Symbiosis
*/
package config

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/symbiosis-cloud/symbiosis-go"
)

// createCmd represents the create command
var Init = &cobra.Command{
	Use:   "init <api_key>",
	Short: "Intialises configuration for the Symbiosis CLI",
	Long:  ``,
	RunE: func(command *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("Please provide an API Key to sym config init")
		}

		apiKey := args[0]

		client, err := symbiosis.NewClientFromAPIKey(apiKey)

		if err != nil {
			return err
		}

		// fire a test call
		_, err = client.Cluster.List(1, 1)

		if err != nil {
			return err
		}

		fmt.Println("Successfully initialised")

		viper.Set("auth.api_key", apiKey)
		viper.Set("auth.method", "api_key")
		viper.WriteConfig()

		return nil
	},
}

func init() {
}
