/*
Copyright © 2022 Symbiosis
*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"time"

	_ "embed"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/symbiosis-cloud/cli/cmd/run"
	"github.com/symbiosis-cloud/cli/pkg/command"
	"github.com/symbiosis-cloud/cli/pkg/identity"
	"github.com/symbiosis-cloud/cli/pkg/util"
	"github.com/symbiosis-cloud/symbiosis-go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO: WRITE STATE
type RunState struct {
	ID      string `json:"id"`
	Cluster string `json:"cluster"`
	State   string `json:"state"`
}

type RunCommand struct {
	Client      *symbiosis.Client
	CommandOpts *command.CommandOpts
}

func (c *RunCommand) Execute(cmd *cobra.Command, args []string) error {
	file, err := cmd.Flags().GetString("file")

	if err != nil {
		return err
	}

	kubeConfig, err := cmd.Flags().GetString("identityOutputPath")

	if err != nil {
		return err
	}

	runFile, err := run.ReadRunFile(file, c.CommandOpts)

	if err != nil {
		return err
	}

	err = runFile.RunBuilders()

	if err != nil {
		return err
	}

	// TODO: allow to set id to branch name?
	id, err := uuid.NewUUID()

	if err != nil {
		return err
	}
	clusterName := fmt.Sprintf("run-%s", id.String())

	log.Printf("Creating cluster: %s", clusterName)

	_, err = c.Client.Cluster.Create(&symbiosis.ClusterInput{
		Name:              clusterName,
		Nodes:             []symbiosis.ClusterNodeInput{},
		IsHighlyAvailable: false,
		Region:            "germany-1", // TODO: allow setting custom region
		KubeVersion:       "latest",    // TODO: allow setting custom kubeVersion
	})

	if err != nil {
		return err
	}

	log.Println("Cluster created, Creating identity...")

	identity, err := identity.NewClusterIdentity(c.Client, clusterName, kubeConfig)

	if err != nil {
		return err
	}

	log.Printf("Written identity to %s\n", identity.KubeConfigPath)

	runFile.SetIdentity(identity)

	clientset, err := util.GetKubernetesClient(identity.KubeConfigPath)

	numNodes := 2

	_, err = c.Client.NodePool.Create(&symbiosis.NodePoolInput{
		Name:         fmt.Sprintf("%s-autopool", clusterName),
		ClusterName:  clusterName,
		NodeTypeName: "general-1", // TODO: allow changing default node type
		Quantity:     numNodes,    // TODO: allow setting number of nodes
		Autoscaling: symbiosis.AutoscalingSettings{
			Enabled: true,
			MinSize: 2,
			MaxSize: 10,
		}, // TODO: allow changing autoscaling
		Labels: []symbiosis.NodeLabel{{
			Key:   "managed-by",
			Value: "sym-cli",
		}},
		Taints: []symbiosis.NodeTaint{},
	})

	if err != nil {
		return err
	}

	log.Println("Cluster created, waiting for node pools to become active...")

	it := 0
	for {

		readyNodes := 0

		nodes, _ := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})

		for _, node := range nodes.Items {
		inner:
			for _, condition := range node.Status.Conditions {
				if condition.Type == "Ready" && condition.Status == "True" {
					readyNodes++
					break inner
				}
			}
		}

		if numNodes == readyNodes {
			break
		}

		if it >= 60 {
			return fmt.Errorf("Timout trying to check if new cluster is ready")
		}

		time.Sleep(time.Second * 5)
		it++
	}

	log.Println("Cluster ready for use")

	// run helm
	if runFile.Deploy != nil && runFile.Deploy.Helm != nil {
		log.Println("Installing helm chart...")

		err = runFile.RunDeploy()

		if err != nil {
			return err
		}
		return nil
	}

	log.Println("Run finished.")

	return nil
}

func (c *RunCommand) Command() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run steps defined in sym.yaml",
		Long:  ``,
		RunE:  c.Execute,
	}

	cmd.Flags().String("file", "sym.yaml", "File to use (default: sym.yaml)")
	cmd.Flags().String("identityOutputPath", "", "Write the generated kubeConfig file to this location")

	return cmd
}

func (c *RunCommand) Init(client *symbiosis.Client, opts *command.CommandOpts) {
	c.Client = client
	c.CommandOpts = opts
}
