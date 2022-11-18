package secrets

import "github.com/symbiosis-cloud/symbiosis-go"

type SecretsManager struct {
	client      *symbiosis.Client
	project     string
	environment symbiosis.ProjectEnvironment
}

func (m *SecretsManager) GetSecrets() (symbiosis.SecretCollection, error) {
	return m.client.Secret.GetSecretsByProjectAndEnvironment(m.project, m.environment)
}

func NewSecretManager(client *symbiosis.Client, project string, environment symbiosis.ProjectEnvironment)
