package builder

import (
	"fmt"
	"log"
	"os/exec"
	"path"
	"strings"

	"github.com/symbiosis-cloud/cli/pkg/command"
	"github.com/symbiosis-cloud/cli/pkg/identity"
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
	CommandOpts *command.CommandOpts
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
			log.Printf("Adding repository %s", deployment.Repository.Name)
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

			if b.CommandOpts.Debug {
				log.Println(string(addOutput))
				log.Println(string(updateOutput))
			}
		}
	}

	return nil
}

func (b *HelmBuilder) Deploy() error {
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

	log.Printf("Installing Helm chart %s\n", d.Name)
	kubeConfig := b.GetIdentity().KubeConfigPath

	// check if the helm chart is already installed
	releaseExists := true

	existArgs := []string{"status", "--kubeconfig", kubeConfig, d.Name}
	existsCmd := exec.Command("helm", existArgs...)
	existsOutput, err := existsCmd.CombinedOutput()
	if b.CommandOpts.Debug {
		log.Printf("Executing: helm %s\n", strings.Join(existArgs, " "))
	}

	if err != nil {
		if !strings.Contains(string(existsOutput), "not found") {
			return err
		}

		releaseExists = false
	}

	if b.CommandOpts.Debug {
		fmt.Println(err)
		fmt.Println(string(existsOutput))
	}

	chart := d.Chart

	if d.Repository == nil {
		chart = b.expandPaths(d.Chart)
	}

	var args []string

	if releaseExists {
		log.Printf("Release %s already exists, upgrading...", d.Name)
		args = []string{"upgrade", "--kubeconfig", kubeConfig}
	} else {
		args = []string{"install", "--kubeconfig", kubeConfig, d.Name, chart}
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

	log.Printf("Running helm %s\n", strings.Join(args, " "))

	repoAdd := exec.Command("helm", args...)
	output, err := repoAdd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Helm failed. Full output: %s", output)
	}

	if b.CommandOpts.Debug {
		log.Println(string(output))
	}

	return nil
}

func (b *HelmBuilder) requirements() ([]Requirement, error) {
	requirements := []Requirement{
		&CommandRequirement{"helm"},
	}

	for _, d := range b.deployments {
		if d.ValuesFile != "" {
			log.Printf("Using values file %s", d.ValuesFile)
			requirements = append(requirements, &FileRequirement{b.expandPaths(d.ValuesFile)})
		}

		chartFile := path.Join(b.expandPaths(d.Chart), "Chart.yaml")
		chartExists := FileExists(chartFile)

		if !chartExists && d.Repository == nil {
			return nil, fmt.Errorf("Chart not found in path %s", chartFile)
		} else if chartExists && d.Repository != nil {
			log.Println("[Warning] Supplied a local chart and a repository, ignoring repository")
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

func NewHelmBuilder(deployments []HelmDeployment, dir string, opts *command.CommandOpts) *HelmBuilder {
	log.Println("Using Helm for deployment")
	return &HelmBuilder{
		deployments: deployments,
		dir:         dir,
		CommandOpts: opts,
	}
}
