package identity

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"text/template"

	_ "embed"

	"github.com/symbiosis-cloud/symbiosis-go"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdlatest "k8s.io/client-go/tools/clientcmd/api/latest"

	"sigs.k8s.io/yaml"
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
	Output         []byte
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

func NewClusterIdentity(client *symbiosis.Client, clusterName string, kubeConfigPath string, merge bool) (*ClusterIdentity, error) {
	var outputFile *os.File

	if kubeConfigPath != "" {

		file, err := os.OpenFile(kubeConfigPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		defer file.Close()

		if err != nil {
			return nil, err
		}

		outputFile = file
	} else {

		file, err := ioutil.TempFile(os.TempDir(), "symbiosis")
		defer file.Close()
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

	output, err := os.ReadFile(outputFile.Name())

	if err != nil {
		return nil, err
	}

	if merge {
		rules := clientcmd.NewDefaultClientConfigLoadingRules()
		rules.Precedence = append(rules.Precedence, outputFile.Name())
		mergedConfig, err := rules.Load()

		if err != nil {
			return nil, err
		}
		json, err := runtime.Encode(clientcmdlatest.Codec, mergedConfig)
		if err != nil {
			fmt.Printf("Unexpected error: %v", err)
		}

		mergedOutput, err := yaml.JSONToYAML(json)
		if err != nil {
			fmt.Printf("Unexpected error: %v", err)
		}

		output = mergedOutput

		err = ioutil.WriteFile(clientcmd.RecommendedHomeFile, mergedOutput, 0644)
	}

	return &ClusterIdentity{
		KubeConfigPath: outputFile.Name(),
		KubeConfig:     kubeConfig,
		Output:         output,
	}, nil
}
