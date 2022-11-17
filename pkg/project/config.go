package project

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path"
	"path/filepath"
	"reflect"

	"github.com/manifoldco/promptui"
	"github.com/symbiosis-cloud/cli/pkg/builder"
	"github.com/symbiosis-cloud/cli/pkg/identity"
	"github.com/symbiosis-cloud/cli/pkg/symcommand"
	"github.com/symbiosis-cloud/symbiosis-go"
	"gopkg.in/yaml.v2"
)

type Deployment struct {
	Helm      []builder.HelmDeployment `yaml:"helm,omitempty"`
	Kustomize []struct {
		Path string `yaml:"path"`
	} `yaml:"kustomize,omitempty"`
}

type ProjectConfig struct {
	Project *symbiosis.Project
	Deploy  *Deployment `yaml:"deploy"`
	Test    []struct {
		Deployment []struct {
			Name    string `yaml:"name"`
			Command string `yaml:"command"`
		} `yaml:"deployment,omitempty"`
		Image   string `yaml:"image,omitempty"`
		Command string `yaml:"command,omitempty"`
	} `yaml:"test,omitempty"`
	Preview struct {
	} `yaml:"preview"`

	builders        []builder.Builder
	Path            string
	rawConfig       []byte
	client          *symbiosis.Client
	commandOpts     *symcommand.CommandOpts
	ProjectFilePath string
}

func (p *ProjectConfig) Parse() error {

	if _, err := os.Stat(p.Path); os.IsNotExist(err) {
		return fmt.Errorf("Run config file %s not found", p.Path)
	}
	f, err := os.ReadFile(p.Path)

	if err != nil {
		return err
	}

	secrets, err := p.client.Secret.GetSecretsByProject(p.Project.Name)

	if err != nil {
		return err
	}

	parsedFile := bytes.NewBuffer([]byte{})

	t := template.Must(template.New("parse-project-config").Funcs(template.FuncMap{
		"Secret": func(secretName string) (string, error) {
			if secret := secrets[secretName]; secret != nil {
				return secret.Value, nil
			}
			return "", fmt.Errorf("Secret %s could not be found in project %s", secretName, p.Project.Name)
		},
	}).Parse(string(f)))

	err = t.Execute(parsedFile, nil)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(parsedFile.Bytes(), &p)

	if err != nil {
		return err
	}

	if p.Deploy == nil {
		return fmt.Errorf("No deployments configured... cannot continue")
	}

	if p.Deploy.Helm != nil {
		helm := builder.NewHelmBuilder(p.Deploy.Helm, filepath.Dir(p.Path), p.commandOpts)
		p.builders = append(p.builders, helm)
	}

	// store the raw config for further processing
	p.rawConfig = f

	return nil
}

func (p *ProjectConfig) SetIdentity(identity *identity.ClusterIdentity) {
	for _, builder := range p.builders {
		builder.SetIdentity(identity)
	}
}

func (p *ProjectConfig) RunBuilders() error {
	for _, builder := range p.builders {
		err := builder.Build()

		if err != nil {
			return err
		}
	}

	return nil
}

func (p *ProjectConfig) RunDeploy() error {
	for _, builder := range p.builders {
		err := builder.Deploy()

		if err != nil {
			return err
		}
	}

	return nil
}

func (p *ProjectConfig) PromptProject(path string) (*symbiosis.Project, error) {

	projects, err := p.client.Project.List()

	if err != nil {
		return nil, err
	}

	if len(projects) == 0 {
		return nil, fmt.Errorf("No projects found. Please create a project first.")
	}

	projectList := make(map[string]*symbiosis.Project, len(projects))

	for _, project := range projects {
		projectList[project.Name] = project
	}

	prompt := promptui.Select{
		Label: "Select Project",
		Items: reflect.ValueOf(projectList).MapKeys(),
	}

	_, result, err := prompt.Run()

	if err != nil {
		p.commandOpts.Logger.Error().Msgf("Failed to select project %v", err)
		return nil, err
	}

	p.commandOpts.Logger.Info().Msgf("Project selected: %q", result)

	project := projectList[result]

	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer file.Close()

	output, err := json.Marshal(project)

	if err != nil {
		return nil, err
	}

	_, err = file.Write(output)

	if err != nil {
		return nil, err
	}

	return project, nil
}

func NewProjectConfig(file string, opts *symcommand.CommandOpts, client *symbiosis.Client) (*ProjectConfig, error) {
	filePath, err := filepath.Abs(file)

	dir, err := os.Getwd()

	if err != nil {
		return nil, err
	}

	var project *symbiosis.Project

	projectConfig := &ProjectConfig{
		Path:        filePath,
		client:      client,
		commandOpts: opts,
	}

	if opts.Project != nil {
		project = opts.Project
	} else {
		projectFilePath := path.Join(dir, ".symbiosis.project")
		projectConfig.ProjectFilePath = projectFilePath

		projectFile, err := os.ReadFile(projectFilePath)

		if err != nil && !os.IsNotExist(err) {
			return nil, err
		} else if err != nil && os.IsNotExist(err) {
			p, err := projectConfig.PromptProject(projectFilePath)

			if err != nil {
				return nil, err
			}

			project = p
		} else {
			err = json.Unmarshal(projectFile, &project)

			if err != nil {
				return nil, err
			}

		}
	}

	projectConfig.Project = project

	return projectConfig, nil
}
