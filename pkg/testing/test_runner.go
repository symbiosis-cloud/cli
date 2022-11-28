package testing

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/symbiosis-cloud/cli/pkg/identity"
	"github.com/symbiosis-cloud/cli/pkg/output"
	"github.com/symbiosis-cloud/cli/pkg/symcommand"
	"golang.org/x/sync/errgroup"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubernetes "k8s.io/client-go/kubernetes"
)

type TestState string

const (
	TEST_STATE_FAILED  TestState = "FAILED"
	TEST_STATE_SUCCESS TestState = "SUCCESS"
	TEST_STATE_PENDING TestState = "PENDING"
)

type TestJob struct {
	Image    string    `json:"image"`
	Commands []string  `json:"commands"`
	State    TestState `json:"state"`
}

type TestResult struct {
	Image    string    `json:"image"`
	Commands []string  `json:"commands"`
	State    TestState `json:"state"`
	ExitCode int32     `json:"exitCode"`
	Logs     string    `json:"logs"`
}

type TestRunner struct {
	jobs []*TestJob

	identity  *identity.ClusterIdentity
	clientSet *kubernetes.Clientset

	CommandOpts *symcommand.CommandOpts
}

func (t *TestRunner) Run() error {

	t.CommandOpts.Logger.Info().Msgf("Running %d tests...", len(t.jobs))

	// TODO: make timeouts configurable
	ctx, cancel := context.WithTimeout(context.TODO(), time.Minute*30)

	var results []*TestResult

	defer cancel()

	errGroup := new(errgroup.Group)
	podsApi := t.clientSet.CoreV1().Pods(t.CommandOpts.Namespace)

	var deletePods []string

	for i, job := range t.jobs {

		podName := fmt.Sprintf("test-job-%d", i)
		deletePods = append(deletePods, podName)

		podSpec := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      podName,
				Namespace: t.CommandOpts.Namespace,
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:    fmt.Sprintf("test-container-%d", i),
						Image:   job.Image,
						Command: job.Commands,
					},
				},
				RestartPolicy: v1.RestartPolicyNever,
			},
		}

		runningJob := job

		errGroup.Go(func() error {
			_, err := podsApi.Create(ctx, podSpec, metav1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("Failed to create pod for test %d", i)
			}

			result := make(chan *TestResult, 1)
			errchan := make(chan error, 1)

			defer close(result)
			defer close(errchan)

			quit := make(chan bool)

			for {

				go func() {

					for {

						select {
						case <-quit:
							return
						default:

							pod, err := podsApi.Get(ctx, podName, metav1.GetOptions{})

							if err != nil {
								errchan <- err
								quit <- true
							}

							for _, x := range pod.Status.ContainerStatuses {
								if x.State.Terminated != nil && x.State.Terminated.Reason != "" {

									req := podsApi.GetLogs(pod.Name, &v1.PodLogOptions{})
									podLogs, err := req.Stream(ctx)
									if err != nil {
										errchan <- err
										quit <- true
									}
									defer podLogs.Close()

									buf := new(bytes.Buffer)
									_, err = io.Copy(buf, podLogs)
									if err != nil {
										errchan <- err
										quit <- true
									}

									logs := buf.String()

									t.CommandOpts.Logger.Debug().Msgf("Container logs: %s", logs)
									var state TestState

									if x.State.Terminated.Reason == "Completed" {
										state = TEST_STATE_SUCCESS
									} else {
										state = TEST_STATE_FAILED
									}

									result <- &TestResult{
										Image:    runningJob.Image,
										Commands: runningJob.Commands,
										State:    state,
										ExitCode: x.State.Terminated.ExitCode,
										Logs:     logs,
									}
									quit <- true
								}
							}

							time.Sleep(5)
						}
					}
				}()

				select {
				case res := <-result:
					results = append(results, res)
					return nil
				case err := <-errchan:
					return err
				case <-time.After(time.Minute * 30):
					return fmt.Errorf("Timeout occurred")
				}
			}
		})

	}

	err := errGroup.Wait()

	if err != nil {
		return err
	}

	if len(deletePods) > 0 {
		for _, d := range deletePods {
			err := podsApi.Delete(ctx, d, metav1.DeleteOptions{})

			if err != nil {
				return err
			}

			t.CommandOpts.Logger.Debug().Msgf("Cleaned up pod %s", d)
		}
	}

	var data [][]interface{}

	for _, result := range results {
		data = append(data, []interface{}{result.Image, result.Commands, result.State, result.ExitCode})
	}

	err = output.NewOutput(output.TableOutput{
		Headers: []string{"Test image", "Command", "State", "Exit code"},
		Data:    data,
	},
		results,
	).VariableOutput()

	if err != nil {
		return err
	}

	return nil

}

func NewTestRunner(jobs []*TestJob, clientSet *kubernetes.Clientset, opts *symcommand.CommandOpts) (*TestRunner, error) {
	return &TestRunner{
		jobs:        jobs,
		CommandOpts: opts,
		clientSet:   clientSet,
	}, nil
}

func NewTestJob(image string, commands []string) *TestJob {
	return &TestJob{image, commands, TEST_STATE_PENDING}
}
