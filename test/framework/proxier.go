package framework

import (
	"fmt"
	"time"

	maegusv1 "github.com/draveness/proxier/pkg/apis/maegus/v1"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

func (f *Framework) CreateProxierAndWaitUntilReady(ns string, p *maegusv1.Proxier) (*maegusv1.Proxier, error) {
	result, err := f.MaegusClientV1.Proxiers(ns).Create(p)
	if err != nil {
		return nil, fmt.Errorf("creating roxier instances failed (%v): %v", p.Name, err)
	}

	if err := f.WaitForProxierReady(result, 5*time.Minute); err != nil {
		return nil, fmt.Errorf("waiting for Proxier instances timed out (%v): %v", p.Name, err)
	}

	return result, nil
}

func (f *Framework) WaitForProxierReady(p *maegusv1.Proxier, timeout time.Duration) error {
	var pollErr error

	err := wait.Poll(2*time.Second, timeout, func() (bool, error) {
		_, pollErr = f.MaegusClientV1.Proxiers(v1.NamespaceAll).Get(p.Name, metav1.GetOptions{})

		fmt.Println(pollErr)
		if pollErr != nil {
			return false, nil
		}

		return true, nil
	})
	return errors.Wrapf(pollErr, "waiting for Proxier %v/%v: %v", p.Namespace, p.Name, err)
}
