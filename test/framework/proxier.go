package framework

import (
	"fmt"
	"time"

	maegusv1 "github.com/draveness/proxier/pkg/apis/maegus/v1beta1"
	"github.com/draveness/proxier/pkg/controller/proxier"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

// MakeBasicProxier returns a proxier with given versions and weights.
func MakeBasicProxier(ns, name string, versions []string, weights []int32) *maegusv1.Proxier {
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

// CreateProxier creates a proxier.
func (f *Framework) CreateProxier(ns string, p *maegusv1.Proxier) (*maegusv1.Proxier, error) {
	result, err := f.MaegusClientV1.Proxiers(ns).Create(p)
	if err != nil {
		return nil, fmt.Errorf("creating proxier instances failed (%v): %v", p.Name, err)
	}

	return result, nil
}

// CreateProxierAndWaitUntilReady creates a proxier instance and waits until ready.
func (f *Framework) CreateProxierAndWaitUntilReady(ns string, p *maegusv1.Proxier) (*maegusv1.Proxier, error) {
	result, err := f.CreateProxier(ns, p)
	if err != nil {
		return nil, err
	}

	if err := f.WaitForProxierReady(result, 15*time.Second); err != nil {
		return nil, fmt.Errorf("waiting for Proxier instances timed out (%v): %v", p.Name, err)
	}

	return result, nil
}

// UpdateProxier updates a proxier instance.
func (f *Framework) UpdateProxier(ns string, p *maegusv1.Proxier) (*maegusv1.Proxier, error) {
	result, err := f.MaegusClientV1.Proxiers(ns).Update(p)
	if err != nil {
		return nil, fmt.Errorf("updating proxier instances failed (%v): %v", p.Name, err)
	}

	return result, nil
}

// UpdateProxierAndWaitUntilReady updates a proxier instance and waits until ready.
func (f *Framework) UpdateProxierAndWaitUntilReady(ns string, p *maegusv1.Proxier) (*maegusv1.Proxier, error) {
	result, err := f.UpdateProxier(ns, p)
	if err != nil {
		return nil, err
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
		instance, pollErr := f.MaegusClientV1.Proxiers(p.Namespace).Get(p.Name, metav1.GetOptions{})

		if pollErr != nil {
			return false, nil
		}

		deploymentName := proxier.NewDeploymentName(p)
		if err := f.WaitForDeployment(p.Namespace, deploymentName, timeout); err != nil {
			return false, err
		}

		if err := f.WaitForServiceReady(p.Namespace, p.Name, timeout); err != nil {
			return false, err
		}

		for _, backend := range p.Spec.Backends {
			// backendServiceName := fmt.Sprintf("%s-%s-backend", p.Name, backend.Name)
			if err := f.WaitForServiceReady(p.Namespace, fmt.Sprintf("%s-%s-backend", p.Name, backend.Name), timeout); err != nil {
				return false, err
			}
		}

		if instance.Status.ActiveBackends != int32(len(p.Spec.Backends)) {
			return false, nil
		}

		if instance.Status.ObsoleteBackends != 0 {
			return false, nil
		}

		// TODO: check the proxier expected backends with current backends

		return true, nil
	})
	return errors.Wrapf(pollErr, "waiting for Proxier %v/%v: %v", p.Namespace, p.Name, err)
}
