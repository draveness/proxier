package framework

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
)

func CreateClusterRole(kubeClient kubernetes.Interface, relativePath string) error {
	clusterRole, err := parseClusterRoleYaml(relativePath)
	if err != nil {
		return err
	}

	_, err = kubeClient.RbacV1().ClusterRoles().Get(clusterRole.Name, metav1.GetOptions{})

	if err == nil {
		// ClusterRole already exists -> Update
		_, err = kubeClient.RbacV1().ClusterRoles().Update(clusterRole)
		if err != nil {
			return err
		}

	} else {
		// ClusterRole doesn't exists -> Create
		_, err = kubeClient.RbacV1().ClusterRoles().Create(clusterRole)
		if err != nil {
			return err
		}
	}

	return nil
}

func DeleteClusterRole(kubeClient kubernetes.Interface, relativePath string) error {
	clusterRole, err := parseClusterRoleYaml(relativePath)
	if err != nil {
		return err
	}

	return kubeClient.RbacV1().ClusterRoles().Delete(clusterRole.Name, &metav1.DeleteOptions{})
}

func parseClusterRoleYaml(relativePath string) (*rbacv1.ClusterRole, error) {
	manifest, err := PathToOSFile(relativePath)
	if err != nil {
		return nil, err
	}

	clusterRole := rbacv1.ClusterRole{}
	if err := yaml.NewYAMLOrJSONDecoder(manifest, 100).Decode(&clusterRole); err != nil {
		return nil, err
	}

	return &clusterRole, nil
}
