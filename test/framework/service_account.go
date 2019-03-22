package framework

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
)

func CreateServiceAccount(kubeClient kubernetes.Interface, namespace string, relativPath string) (finalizerFn, error) {
	finalizerFn := func() error { return DeleteServiceAccount(kubeClient, namespace, relativPath) }

	serviceAccount, err := parseServiceAccountYaml(relativPath)
	if err != nil {
		return finalizerFn, err
	}
	serviceAccount.Namespace = namespace
	_, err = kubeClient.CoreV1().ServiceAccounts(namespace).Create(serviceAccount)
	if err != nil {
		return finalizerFn, err
	}

	return finalizerFn, nil
}

func parseServiceAccountYaml(relativPath string) (*v1.ServiceAccount, error) {
	manifest, err := PathToOSFile(relativPath)
	if err != nil {
		return nil, err
	}

	serviceAccount := v1.ServiceAccount{}
	if err := yaml.NewYAMLOrJSONDecoder(manifest, 100).Decode(&serviceAccount); err != nil {
		return nil, err
	}

	return &serviceAccount, nil
}

func DeleteServiceAccount(kubeClient kubernetes.Interface, namespace string, relativPath string) error {
	serviceAccount, err := parseServiceAccountYaml(relativPath)
	if err != nil {
		return err
	}

	return kubeClient.CoreV1().ServiceAccounts(namespace).Delete(serviceAccount.Name, nil)
}
