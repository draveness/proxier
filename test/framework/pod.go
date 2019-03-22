package framework

import (
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
