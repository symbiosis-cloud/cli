package testing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"path"
	"time"

	"github.com/symbiosis-cloud/cli/pkg/identity"
	"github.com/symbiosis-cloud/cli/pkg/output"
	"github.com/symbiosis-cloud/cli/pkg/symcommand"
	"golang.org/x/sync/errgroup"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubernetes "k8s.io/client-go/kubernetes"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
)

type TestState string

const (
	TEST_STATE_FAILED  TestState = "FAILED"
	TEST_STATE_SUCCESS TestState = "SUCCESS"
	TEST_STATE_PENDING TestState = "PENDING"
)

type TestJob struct {
	Image          string    `json:"image"`
	Commands       []string  `json:"commands"`
	State          TestState `json:"state"`
	result         *TestResult
	executionStart time.Time
	podsApi        v1core.PodInterface
	*symcommand.CommandOpts
	context.Context
}

type TestResult struct {
	Name      string        `json:"name"`
	Image     string        `json:"image"`
	Commands  []string      `json:"commands"`
	State     TestState     `json:"state"`
	ExitCode  int32         `json:"exitCode"`
	Logs      string        `json:"logs"`
	Duration  time.Duration `json:"duration_s"`
	outputDir string
}

type TestRunner struct {
	jobs        []*TestJob
	identity    *identity.ClusterIdentity
	clientSet   *kubernetes.Clientset
	CommandOpts *symcommand.CommandOpts
}

func (t *TestRunner) Run(testOutputDir string) error {

	t.CommandOpts.Logger.Info().Msgf("Running %d tests...", len(t.jobs))
	t.CommandOpts.Logger.Info().Msgf("Results are being written to %s...", testOutputDir)

	// TODO: make timeouts configurable
	ctx, cancel := context.WithTimeout(context.TODO(), time.Minute*30)

	results := make([]*TestResult, len(t.jobs))

	defer cancel()

	errGroup := new(errgroup.Group)
	podsApi := t.clientSet.CoreV1().Pods(t.CommandOpts.Namespace)

	var deletePods []string

	for i, job := range t.jobs {

		podName := fmt.Sprintf("test-job-%d", i)
		deletePods = append(deletePods, podName)
		job.result = &TestResult{
			Name:      fmt.Sprintf("test-%d", i),
			Image:     job.Image,
			Commands:  job.Commands,
			State:     TEST_STATE_PENDING,
			outputDir: testOutputDir,
		}

		err := job.result.Write()

		if err != nil {
			return err
		}

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

		num := i

		job.executionStart = time.Now()
		job.Context = ctx
		job.podsApi = podsApi
		job.CommandOpts = t.CommandOpts

		runningJob := job

		errGroup.Go(func() error {
			_, err := podsApi.Create(ctx, podSpec, metav1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("Failed to create pod for test %d", num)
			}

			// result := make(chan *TestResult, 1)
			errchan := make(chan error, 1)

			// defer close(result)
			defer close(errchan)

			quit := make(chan bool)

			for {

				go runningJob.Run(podName, quit, errchan)

				select {
				case <-quit:
					err := runningJob.result.Write()

					if err != nil {
						return err
					}
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

	for i, job := range t.jobs {
		data = append(data, []interface{}{job.result.Image, job.result.Commands, job.result.State, job.result.ExitCode, fmt.Sprintf("%s", job.result.Duration.Round(time.Second).String())})
		results[i] = job.result
	}

	err = output.NewOutput(output.TableOutput{
		Headers: []string{"Test image", "Command", "State", "Exit code", "Duration"},
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
	return &TestJob{image, commands, TEST_STATE_PENDING, nil, time.Now(), nil, nil, nil}
}

// make sure we format duration as float of seconds
func (t *TestResult) MarshalJSON() (b []byte, err error) {
	type Alias TestResult
	return json.Marshal(&struct {
		Duration float64 `json:"duration_s"`
		*Alias
	}{
		Duration: t.Duration.Seconds(),
		Alias:    (*Alias)(t),
	})
}

func (t *TestResult) Write() error {
	data, err := json.MarshalIndent(t, "", "  ")

	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path.Join(t.outputDir, fmt.Sprintf("%s.json", t.Name)), data, 0644)

	if err != nil {
		return err
	}

	return nil
}

func (j *TestJob) Run(
	podName string,
	quit chan bool,
	errchan chan error) {
	for {

		select {
		case <-quit:
			return
		default:

			pod, err := j.podsApi.Get(j.Context, podName, metav1.GetOptions{})

			if err != nil {
				errchan <- err
				quit <- true
			}

			for _, x := range pod.Status.ContainerStatuses {
				if x.State.Terminated != nil && x.State.Terminated.Reason != "" {

					req := j.podsApi.GetLogs(pod.Name, &v1.PodLogOptions{})
					podLogs, err := req.Stream(j.Context)
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

					j.CommandOpts.Logger.Debug().Msgf("Container logs: %s", logs)
					var state TestState

					if x.State.Terminated.Reason == "Completed" {
						state = TEST_STATE_SUCCESS
					} else {
						state = TEST_STATE_FAILED
					}

					j.result.State = state
					j.result.ExitCode = x.State.Terminated.ExitCode
					j.result.Logs = logs
					j.result.Duration = time.Now().Sub(j.executionStart)

					quit <- true
				}
			}

			j.result.Duration = time.Now().Sub(j.executionStart)
			j.result.Write()
		}

		// avoid spamming the k8s API
		time.Sleep(time.Second * 1)
	}
}
