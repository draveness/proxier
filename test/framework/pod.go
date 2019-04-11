package framework

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func MakePod(pathToYaml string) (*v1.Pod, error) {
	manifest, err := PathToOSFile(pathToYaml)
	if err != nil {
		return nil, err
	}
	tectonicPromOp := v1.Pod{}
	if err := yaml.NewYAMLOrJSONDecoder(manifest, 100).Decode(&tectonicPromOp); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to decode file %s", pathToYaml))
	}

	return &tectonicPromOp, nil
}

func (f *Framework) CreatePod(namespace string, pod *v1.Pod) error {
	pod.Namespace = namespace
	_, err := f.KubeClient.CoreV1().Pods(namespace).Create(pod)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to create pod %s", pod.Name))
	}
	return nil
}

func (f *Framework) CreatePodAndWaitUntilReady(namespace string, pod *v1.Pod) error {
	if err := f.CreatePod(namespace, pod); err != nil {
		return err
	}

	if err := f.WaitForPodReady(pod, 30*time.Second); err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to create pod %s", pod.Name))
	}

	return nil
}

// GetPodRestartCount returns a map of container names and their restart counts for
// a given pod.
func (f *Framework) GetPodRestartCount(ns, podName string) (map[string]int32, error) {
	pod, err := f.KubeClient.CoreV1().Pods(ns).Get(podName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve pod to get restart count")
	}

	restarts := map[string]int32{}

	for _, status := range pod.Status.ContainerStatuses {
		restarts[status.Name] = status.RestartCount
	}

	return restarts, nil
}

// WaitForPodReady returns when pod shifted to running phase or timeout.
func (f *Framework) WaitForPodReady(pod *v1.Pod, timeout time.Duration) error {
	var pollErr error

	err := wait.Poll(2*time.Second, timeout, func() (bool, error) {
		pod, pollErr := f.KubeClient.CoreV1().Pods(pod.Namespace).Get(pod.Name, metav1.GetOptions{})

		if pollErr != nil {
			return false, nil
		}

		if pod.Status.Phase != v1.PodRunning {
			return false, nil
		}

		return true, nil
	})
	return errors.Wrapf(pollErr, "waiting for Proxier %v/%v: %v", pod.Namespace, pod.Name, err)
}
