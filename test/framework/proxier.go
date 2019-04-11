package framework

import (
	"fmt"
	"time"

	maegusv1 "github.com/draveness/proxier/pkg/apis/maegus/v1"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

// MakeBasicProxier returns a proxier with given versions and weights.
func (f *Framework) MakeBasicProxier(ns, name string, versions []string, weights []int32) *maegusv1.Proxier {
	backends := []maegusv1.BackendSpec{}
	for i := range versions {
		backends = append(backends, maegusv1.BackendSpec{
			Name:   versions[i],
			Weight: weights[i],
			Selector: map[string]string{
				"version": versions[i],
			},
		})
	}

	return &maegusv1.Proxier{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: maegusv1.ProxierSpec{
			Selector: map[string]string{
				"app": name,
			},
			Ports: []maegusv1.ProxierPort{
				{
					Name:     "http",
					Protocol: maegusv1.ProtocolTCP,
					Port:     80,
				},
			},
			Backends: backends,
		},
	}
}

// CreateProxierAndWaitUntilReady creates a proxier instance and waits until ready.
func (f *Framework) CreateProxierAndWaitUntilReady(ns string, p *maegusv1.Proxier) (*maegusv1.Proxier, error) {
	result, err := f.MaegusClientV1.Proxiers(ns).Create(p)
	if err != nil {
		return nil, fmt.Errorf("creating roxier instances failed (%v): %v", p.Name, err)
	}

	if err := f.WaitForProxierReady(result, 15*time.Second); err != nil {
		return nil, fmt.Errorf("waiting for Proxier instances timed out (%v): %v", p.Name, err)
	}

	return result, nil
}

// WaitForProxierReady returns when proxier shifted to running phase or timeout.
func (f *Framework) WaitForProxierReady(p *maegusv1.Proxier, timeout time.Duration) error {
	var pollErr error

	err := wait.Poll(2*time.Second, timeout, func() (bool, error) {
		proxier, pollErr := f.MaegusClientV1.Proxiers(p.Namespace).Get(p.Name, metav1.GetOptions{})

		if pollErr != nil {
			return false, nil
		}

		if proxier.Status.Phase != maegusv1.ProxierRunning {
			return false, nil
		}

		return true, nil
	})
	return errors.Wrapf(pollErr, "waiting for Proxier %v/%v: %v", p.Namespace, p.Name, err)
}
