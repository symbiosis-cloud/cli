package util

import (
	"fmt"
	"regexp"

	"github.com/symbiosis-cloud/symbiosis-go"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func GetKubernetesClient(kubeConfig string) (*kubernetes.Clientset, error) {
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		return nil, err
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

func ParseTaintsAndLabels(taints []string, labels []string) ([]symbiosis.NodeTaint, []symbiosis.NodeLabel, error) {
	nodeTaints := make([]symbiosis.NodeTaint, len(taints))
	nodeLabels := make([]symbiosis.NodeLabel, len(labels))

	for i, taint := range taints {

		re := regexp.MustCompile(`([a-z0-9\-\_]+)\=([a-zA-Z0-9\-\_]+)\=(NoSchedule|NoExecute|PreferNoSchedule)`)
		if ok := re.Match([]byte(taint)); !ok {
			return nil, nil, fmt.Errorf("Taint %s could not be parsed. Format: key=value=NoSchedule. Types currently supported are: NoSchedule, NoExecute and PreferNoSchedule.", taint)
		}

		matches := re.FindAllStringSubmatch(taint, 3)
		nodeTaints[i] = symbiosis.NodeTaint{
			Key:    matches[0][1],
			Value:  matches[0][2],
			Effect: symbiosis.SchedulerEffect(matches[0][3]),
		}
	}

	for x, label := range labels {
		re := regexp.MustCompile(`([a-z0-9\-\_]+)\=([a-zA-Z0-9\-\_]+)`)
		if ok := re.Match([]byte(label)); !ok {
			return nil, nil, fmt.Errorf("Label %s could not be parsed. Format: key=value", label)
		}

		matches := re.FindAllStringSubmatch(label, 2)
		nodeLabels[x] = symbiosis.NodeLabel{
			Key:   matches[0][1],
			Value: matches[0][2],
		}
	}
	return nodeTaints, nodeLabels, nil
}
