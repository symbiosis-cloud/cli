/*
Copyright © 2022 Symbiosis
*/
package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/symbiosis-cloud/cli/pkg/symcommand"
	"github.com/symbiosis-cloud/symbiosis-go"
)

type CompletionCommand struct {
	Client      *symbiosis.Client
	CommandOpts *symcommand.CommandOpts
}

var CompletionCommands = []string{"node-types", "regions", "roles"}

func (c *CompletionCommand) Execute(command *cobra.Command, args []string) {
	switch args[0] {
	case "bash":
		command.Root().GenBashCompletion(os.Stdout)
	case "zsh":
		command.Root().GenZshCompletion(os.Stdout)
	case "fish":
		command.Root().GenFishCompletion(os.Stdout, true)
	case "powershell":
		command.Root().GenPowerShellCompletionWithDesc(os.Stdout)
	}

}

func (c *CompletionCommand) Command() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate completion script",
		Long: fmt.Sprintf(`To load completions:

	Bash:

		$ source <(%[1]s completion bash)

		# To load completions for each session, execute once:
		# Linux:
		$ %[1]s completion bash > /etc/bash_completion.d/%[1]s
		# macOS:
		$ %[1]s completion bash > $(brew --prefix)/etc/bash_completion.d/%[1]s

	Zsh:

		# If shell completion is not already enabled in your environment,
		# you will need to enable it.  You can execute the following once:

		$ echo "autoload -U compinit; compinit" >> ~/.zshrc

		# To load completions for each session, execute once:
		$ %[1]s completion zsh > "${fpath[1]}/_%[1]s"

		# You will need to start a new shell for this setup to take effect.

	fish:

		$ %[1]s completion fish | source

		# To load completions for each session, execute once:
		$ %[1]s completion fish > ~/.config/fish/completions/%[1]s.fish

	PowerShell:

		PS> %[1]s completion powershell | Out-String | Invoke-Expression

		# To load completions for every new session, run:
		PS> %[1]s completion powershell > %[1]s.ps1
		# and source this file from your PowerShell profile.
	`, "sym"),
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Run:                   c.Execute,
	}

	symcommand.SetDeploymentFlags(cmd)

	return cmd
}

func (c *CompletionCommand) Init(client *symbiosis.Client, opts *symcommand.CommandOpts) {
	c.Client = client
	c.CommandOpts = opts
}