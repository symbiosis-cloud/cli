package identity

import (
	"encoding/base64"
	"io/ioutil"
	"os"
	"text/template"

	_ "embed"

	"github.com/symbiosis-cloud/symbiosis-go"
)

type KubeConfig struct {
	Endpoint string
	Name     string
	Identity *symbiosis.ClusterIdentity
	File     *os.File
}

type ClusterIdentity struct {
	KubeConfigPath string
	KubeConfig     *KubeConfig
}

//go:embed kubeconfig.tmpl
var kubeConfigTemplate string

func (c *ClusterIdentity) Remove() error {
	err := os.Remove(c.KubeConfig.File.Name())

	if err != nil {
		return err
	}

	return nil
}

func NewClusterIdentity(client *symbiosis.Client, clusterName string, kubeConfigPath string) (*ClusterIdentity, error) {
	var outputFile *os.File

	if kubeConfigPath != "" {

		file, err := os.OpenFile(kubeConfigPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

		if err != nil {
			return nil, err
		}

		outputFile = file
	} else {

		file, err := ioutil.TempFile(os.TempDir(), "symbiosis")
		if err != nil {
			return nil, err
		}

		outputFile = file

	}

	cluster, err := client.Cluster.Describe(clusterName)

	if err != nil {
		return nil, err
	}

	identity, err := client.Cluster.GetIdentity(clusterName)

	if err != nil {
		return nil, err
	}

	kubeConfig := &KubeConfig{
		Endpoint: cluster.APIServerEndpoint,
		Name:     clusterName,
		Identity: &symbiosis.ClusterIdentity{
			ClusterCertificateAuthorityPem: base64.URLEncoding.EncodeToString([]byte(identity.ClusterCertificateAuthorityPem)),
			PrivateKeyPem:                  base64.URLEncoding.EncodeToString([]byte(identity.PrivateKeyPem)),
			CertificatePem:                 base64.URLEncoding.EncodeToString([]byte(identity.CertificatePem)),
		},
	}

	t := template.Must(template.New("kube-config").Parse(kubeConfigTemplate))
	err = t.Execute(outputFile, kubeConfig)
	if err != nil {
		return nil, err
	}

	return &ClusterIdentity{
		KubeConfigPath: outputFile.Name(),
		KubeConfig:     kubeConfig,
	}, nil
}
