package framework

import (
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

// WaitForServiceReady returns when service is ready.
func (f *Framework) WaitForServiceReady(namespace, name string, timeout time.Duration) error {
	return wait.Poll(time.Second, timeout, func() (bool, error) {
		_, err := f.KubeClient.CoreV1().Services(namespace).Get(name, metav1.GetOptions{})

		if err != nil {
			if apierrors.IsNotFound(err) {
				return false, nil
			}

			return false, err
		}

		return true, nil
	})
}

// WaitUntilServiceGone returns when service not found.
func (f *Framework) WaitUntilServiceGone(namespace, name string, timeout time.Duration) error {
	return wait.Poll(time.Second, timeout, func() (bool, error) {
		_, err := f.KubeClient.CoreV1().Services(namespace).Get(name, metav1.GetOptions{})

		if err != nil {
			if apierrors.IsNotFound(err) {
				return true, nil
			}

			return false, err
		}

		return false, nil
	})
}
