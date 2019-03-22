package framework

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
)

func CreateRoleBinding(kubeClient kubernetes.Interface, ns string, relativePath string) (finalizerFn, error) {
	finalizerFn := func() error { return DeleteRoleBinding(kubeClient, ns, relativePath) }
	roleBinding, err := parseRoleBindingYaml(relativePath)
	if err != nil {
		return finalizerFn, err
	}

	_, err = kubeClient.RbacV1().RoleBindings(ns).Create(roleBinding)
	return finalizerFn, err
}

func DeleteRoleBinding(kubeClient kubernetes.Interface, ns string, relativePath string) error {
	roleBinding, err := parseRoleBindingYaml(relativePath)
	if err != nil {
		return err
	}

	return kubeClient.RbacV1().RoleBindings(ns).Delete(roleBinding.Name, &metav1.DeleteOptions{})
}

func parseRoleBindingYaml(relativePath string) (*rbacv1.RoleBinding, error) {
	manifest, err := PathToOSFile(relativePath)
	if err != nil {
		return nil, err
	}

	roleBinding := rbacv1.RoleBinding{}
	if err := yaml.NewYAMLOrJSONDecoder(manifest, 100).Decode(&roleBinding); err != nil {
		return nil, err
	}

	return &roleBinding, nil
}
