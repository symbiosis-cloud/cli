/*
Copyright © 2022 Symbiosis
*/
package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/symbiosis-cloud/cli/pkg/identity"
	"github.com/symbiosis-cloud/cli/pkg/project"
	"github.com/symbiosis-cloud/cli/pkg/symcommand"
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
	CommandOpts *symcommand.CommandOpts
}

var (
	merge bool
)

func (c *RunCommand) Execute(command *cobra.Command, args []string) error {

	deploymentFlags, err := symcommand.GetDeploymentFlags(command)

	clusterName, err := command.Flags().GetString("cluster-name")

	if err != nil {
		return err
	}

	region, err := command.Flags().GetString("region")

	if err != nil {
		return err
	}

	_, err = c.Client.Cluster.Describe(clusterName)

	// cluster does not exist, create it
	if err != nil && strings.Contains(err.Error(), "not found") {
		c.CreateCluster(clusterName, region, deploymentFlags.IdentityOutputPath)
	} else if err != nil {
		return err
	} else {
		c.CommandOpts.Logger.Info().Msgf("Using existing cluster: %s", clusterName)
	}

	c.CommandOpts.Namespace = deploymentFlags.Namespace

	projectConfig, err := project.NewProjectConfig(deploymentFlags.File, c.CommandOpts, c.Client, nil)

	if err != nil {
		return err
	}

	err = projectConfig.Parse()

	if err != nil {
		return err
	}

	err = projectConfig.RunBuilders()

	if err != nil {
		return err
	}

	identity, err := identity.NewClusterIdentity(c.Client, clusterName, deploymentFlags.IdentityOutputPath, merge)

	if err != nil {
		return err
	}

	c.CommandOpts.Logger.Info().Msgf("Written identity to %s", identity.KubeConfigPath)

	projectConfig.SetIdentity(identity)

	// run helm
	if projectConfig.Deploy != nil && projectConfig.Deploy.Helm != nil {
		c.CommandOpts.Logger.Info().Msg("Installing helm chart...")

		err = projectConfig.RunDeploy()

		if err != nil {
			return err
		}
		return nil
	}

	c.CommandOpts.Logger.Info().Msg("Run finished.")

	return nil
}

func (c *RunCommand) Command() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run steps defined in sym.yaml",
		Long:  ``,
		RunE:  c.Execute,
	}

	cmd.Flags().String("region", "germany-1", "Set the Symbiosis region")
	cmd.Flags().String("cluster-name", fmt.Sprintf("run-%s", util.RandomString(8)), "Set the Cluster name of the newly created cluster")

	symcommand.SetDeploymentFlags(cmd)

	return cmd
}

func (c *RunCommand) CreateCluster(clusterName string, region string, outputPath string) error {
	c.CommandOpts.Logger.Info().Msgf("Creating cluster: %s", clusterName)

	// TODO: allow changing node pool size
	numNodes := 2

	_, err := c.Client.Cluster.Create(&symbiosis.ClusterInput{
		Name: clusterName,
		Nodes: []symbiosis.ClusterNodePoolInput{{
			Name:         fmt.Sprintf("%s-autopool", clusterName),
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
		}},
		IsHighlyAvailable: false,
		Region:            region,
		KubeVersion:       "latest", // TODO: allow setting custom kubeVersion
	})

	if err != nil {
		return err
	}

	c.CommandOpts.Logger.Info().Msg("Cluster created, Creating identity...")
	c.CommandOpts.Logger.Info().Msg("Cluster created, waiting for node pools to become active...")

	identity, err := identity.NewClusterIdentity(c.Client, clusterName, outputPath, merge)

	if err != nil {
		return err
	}

	clientset, err := util.GetKubernetesClient(identity.KubeConfigPath)

	if err != nil {
		return err
	}

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

		if readyNodes > 0 {
			break
		}

		if it >= 60 {
			return fmt.Errorf("Timout trying to check if new cluster is ready")
		}

		time.Sleep(time.Second * 10)
		it++
	}

	c.CommandOpts.Logger.Info().Msg("Cluster ready for use")

	return nil
}

func (c *RunCommand) Init(client *symbiosis.Client, opts *symcommand.CommandOpts) {
	c.Client = client
	c.CommandOpts = opts
}
