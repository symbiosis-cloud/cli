package cmd

import (
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
)

func ConfirmPreRunE(prompt string) error {
	force := viper.GetBool("force")

	if merge && !force {
		prompt := promptui.Prompt{
			Label:     prompt,
			IsConfirm: true,
		}

		_, err := prompt.Run()

		if err != nil {
			return fmt.Errorf("User cancelled action")
		}
	}

	return nil
}
