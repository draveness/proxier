package framework

import (
	"fmt"
	"testing"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateNamespace(kubeClient kubernetes.Interface, name string) (*v1.Namespace, error) {
	namespace, err := kubeClient.CoreV1().Namespaces().Create(&v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to create namespace with name %v", name))
	}
	return namespace, nil
}

func (ctx *TestCtx) CreateNamespace(t *testing.T, kubeClient kubernetes.Interface) string {
	name := ctx.GetObjID()
	if _, err := CreateNamespace(kubeClient, name); err != nil {
		t.Fatal(err)
	}

	namespaceFinalizerFn := func() error {
		return DeleteNamespace(kubeClient, name)
	}

	ctx.AddFinalizerFn(namespaceFinalizerFn)

	return name
}

func DeleteNamespace(kubeClient kubernetes.Interface, name string) error {
	return kubeClient.CoreV1().Namespaces().Delete(name, nil)
}

func AddLabelsToNamespace(kubeClient kubernetes.Interface, name string, additionalLabels map[string]string) error {
	ns, err := kubeClient.CoreV1().Namespaces().Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if ns.Labels == nil {
		ns.Labels = map[string]string{}
	}

	for k, v := range additionalLabels {
		ns.Labels[k] = v
	}

	_, err = kubeClient.CoreV1().Namespaces().Update(ns)
	if err != nil {
		return err
	}

	return nil
}
