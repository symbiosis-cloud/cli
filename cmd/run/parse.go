package run

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/symbiosis-cloud/cli/pkg/builder"
	"github.com/symbiosis-cloud/cli/pkg/command"
	"github.com/symbiosis-cloud/cli/pkg/identity"
	"gopkg.in/yaml.v3"
)

type Deployment struct {
	Helm      []builder.HelmDeployment `yaml:"helm,omitempty"`
	Kustomize []struct {
		Path string `yaml:"path"`
	} `yaml:"kustomize,omitempty"`
}

type RunFile struct {
	Deploy *Deployment `yaml:"deploy"`
	Test   []struct {
		Deployment []struct {
			Name    string `yaml:"name"`
			Command string `yaml:"command"`
		} `yaml:"deployment,omitempty"`
		Image   string `yaml:"image,omitempty"`
		Command string `yaml:"command,omitempty"`
	} `yaml:"test,omitempty"`
	Preview struct {
	} `yaml:"preview"`

	builders []builder.Builder
	Path     string
}

func ReadRunFile(file string, opts *command.CommandOpts) (*RunFile, error) {

	if _, err := os.Stat(file); os.IsNotExist(err) {
		return nil, fmt.Errorf("Run config file %s not found", file)
	}
	f, err := os.ReadFile(file)

	if err != nil {
		return nil, err
	}

	var runFile *RunFile

	err = yaml.Unmarshal(f, &runFile)

	if err != nil {
		return nil, err
	}

	if runFile.Deploy == nil {
		return nil, fmt.Errorf("No deployments configured... cannot continue")
	}

	path, err := filepath.Abs(file)

	if err != nil {
		return nil, err
	}

	runFile.Path = path

	if runFile.Deploy.Helm != nil {
		helm := builder.NewHelmBuilder(runFile.Deploy.Helm, filepath.Dir(path), opts)
		runFile.builders = append(runFile.builders, helm)
	}

	return runFile, nil
}

func (r *RunFile) SetIdentity(identity *identity.ClusterIdentity) {
	for _, builder := range r.builders {
		builder.SetIdentity(identity)
	}
}

func (r *RunFile) RunBuilders() error {
	for _, builder := range r.builders {
		err := builder.Build()

		if err != nil {
			return err
		}
	}

	return nil
}

func (r *RunFile) RunDeploy() error {
	for _, builder := range r.builders {
		err := builder.Deploy()

		if err != nil {
			return err
		}
	}

	return nil
}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}
