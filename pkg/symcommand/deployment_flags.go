package symcommand

import "github.com/spf13/cobra"

type DeploymentFlags struct {
	Namespace          string
	IdentityOutputPath string
	File               string
}

func SetDeploymentFlags(command *cobra.Command) {
	command.Flags().String("namespace", "default", "Set the deployment namespace (default: default)")
	command.Flags().String("identity-output-path", "", "Write the generated kubeConfig file to this location")
	command.Flags().String("file", "sym.yaml", "File to use (default: sym.yaml)")
}

func GetDeploymentFlags(command *cobra.Command) (*DeploymentFlags, error) {
	namespace, err := command.Flags().GetString("namespace")

	if err != nil {
		return nil, err
	}

	identityOutputPath, err := command.Flags().GetString("identity-output-path")

	if err != nil {
		return nil, err
	}

	file, err := command.Flags().GetString("file")

	if err != nil {
		return nil, err
	}

	return &DeploymentFlags{
		Namespace:          namespace,
		IdentityOutputPath: identityOutputPath,
		File:               file,
	}, nil

}
