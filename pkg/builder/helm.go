package builder

import (
	"fmt"
	"os/exec"
	"path"
	"strings"

	"github.com/symbiosis-cloud/cli/pkg/identity"
	"github.com/symbiosis-cloud/cli/pkg/symcommand"
	"golang.org/x/sync/errgroup"
)

type HelmRepository struct {
	Name string `yaml:"name"`
	Url  string `yaml:"url"`
}

type HelmDeployment struct {
	Name       string            `yaml:"name"`
	Chart      string            `yaml:"chart"`
	ValuesFile string            `yaml:"valuesFile"`
	Values     map[string]string `yaml:"values"`
	Repository *HelmRepository   `yaml:"repository"`
}

type HelmBuilder struct {
	identity    *identity.ClusterIdentity
	deployments []HelmDeployment
	dir         string
	CommandOpts *symcommand.CommandOpts
}

func (b *HelmBuilder) GetIdentity() *identity.ClusterIdentity {
	return b.identity
}

func (b *HelmBuilder) SetIdentity(identity *identity.ClusterIdentity) {
	b.identity = identity
}

func (b *HelmBuilder) Build() error {
	err := meetsRequirements(b)

	if err != nil {
		return err
	}

	for _, deployment := range b.deployments {
		if deployment.Repository != nil {
			b.CommandOpts.Logger.Info().Msgf("Adding repository %s", deployment.Repository.Name)
			args := []string{"repo", "add", deployment.Repository.Name, deployment.Repository.Url}

			repoAdd := exec.Command("helm", args...)
			addOutput, err := repoAdd.Output()
			if err != nil {
				return err
			}

			repoUpdate := exec.Command("helm", "repo", "update")
			updateOutput, err := repoUpdate.CombinedOutput()
			if err != nil {
				return err
			}

			b.CommandOpts.Logger.Debug().Msg(string(addOutput))
			b.CommandOpts.Logger.Debug().Msg(string(updateOutput))

		}
	}

	return nil
}

func (b *HelmBuilder) Deploy() error {
	b.CommandOpts.Logger.Info().Msg("Using Helm for deployment")
	w := new(errgroup.Group)

	for _, d := range b.deployments {

		// avoid issues with variable capturing
		deployment := d

		w.Go(func() error {
			err := b.Install(deployment)

			if err != nil {
				return err
			}

			return nil
		})
	}

	if err := w.Wait(); err != nil {
		return err
	}

	return nil
}

func (b *HelmBuilder) Install(d HelmDeployment) error {
	b.CommandOpts.Logger.Info().Msgf("Installing Helm chart %s", d.Name)
	kubeConfig := b.GetIdentity().KubeConfigPath

	namespace := b.CommandOpts.Namespace

	// check if the helm chart is already installed
	releaseExists := true

	existArgs := []string{"status", "--kubeconfig", kubeConfig, d.Name, "--namespace", namespace}
	existsCmd := exec.Command("helm", existArgs...)
	existsOutput, err := existsCmd.CombinedOutput()

	b.CommandOpts.Logger.Debug().Msgf("Executing: helm %s", strings.Join(existArgs, " "))

	if err != nil {
		if !strings.Contains(string(existsOutput), "not found") {
			return err
		}

		releaseExists = false
	}

	b.CommandOpts.Logger.Debug().Err(err)
	b.CommandOpts.Logger.Debug().Msg(string(existsOutput))

	chart := d.Chart

	if d.Repository == nil {
		chart = b.expandPaths(d.Chart)
	}

	var args []string

	if releaseExists {
		b.CommandOpts.Logger.Info().Msgf("Release %s already exists, upgrading...", d.Name)
		args = []string{"upgrade", "--kubeconfig", kubeConfig, "--namespace", namespace}
	} else {
		args = []string{"install", "--kubeconfig", kubeConfig, "--namespace", namespace, d.Name, chart}
	}

	for key, value := range d.Values {
		args = append(args, "--set", fmt.Sprintf("%s=%s", key, value))
	}

	if d.ValuesFile != "" {
		args = append(args, "-f", b.expandPaths(d.ValuesFile))
	}

	if releaseExists {
		args = append(args, d.Name, chart)
	}

	b.CommandOpts.Logger.Debug().Msgf("Running helm %s", strings.Join(args, " "))

	repoAdd := exec.Command("helm", args...)
	output, err := repoAdd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Helm failed. Full output: %s", output)
	}

	b.CommandOpts.Logger.Debug().Msg(string(output))

	return nil
}

func (b *HelmBuilder) requirements() ([]Requirement, error) {
	requirements := []Requirement{
		&CommandRequirement{"helm"},
	}

	for _, d := range b.deployments {
		if d.ValuesFile != "" {
			b.CommandOpts.Logger.Info().Msgf("Using values file %s", d.ValuesFile)
			requirements = append(requirements, &FileRequirement{b.expandPaths(d.ValuesFile)})
		}

		chartFile := path.Join(b.expandPaths(d.Chart), "Chart.yaml")
		chartExists := FileExists(chartFile)

		if !chartExists && d.Repository == nil {
			return nil, fmt.Errorf("Chart not found in path %s", chartFile)
		} else if chartExists && d.Repository != nil {
			b.CommandOpts.Logger.Info().Msg("[Warning] Supplied a local chart and a repository, ignoring repository")
		}

		if d.Repository != nil {
			if d.Repository.Name == "" || d.Repository.Url == "" {
				return nil, fmt.Errorf("Invalid repository configuration for Helm chart")
			}
		}

	}

	return requirements, nil
}

func (b *HelmBuilder) expandPaths(path string) string {
	if strings.Contains(path, "./") {
		return strings.Replace(path, "./", b.dir+"/", 1)
	}

	return path
}

func NewHelmBuilder(deployments []HelmDeployment, dir string, opts *symcommand.CommandOpts) *HelmBuilder {
	return &HelmBuilder{
		deployments: deployments,
		dir:         dir,
		CommandOpts: opts,
	}
}
