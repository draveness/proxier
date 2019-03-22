package framework

import (
	"net/http"
	"time"

	"github.com/coreos/prometheus-operator/pkg/k8sutil"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Framework struct {
	KubeClient     kubernetes.Interface
	HTTPClient     *http.Client
	MasterHost     string
	DefaultTimeout time.Duration
}

// New setups a test framework and returns it.
func New(kubeconfig, opImage string) (*Framework, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, errors.Wrap(err, "build config from flags failed")
	}

	cli, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "creating new kube-client failed")
	}

	httpc := cli.CoreV1().RESTClient().(*rest.RESTClient).Client
	if err != nil {
		return nil, errors.Wrap(err, "creating http-client failed")
	}

	f := &Framework{
		MasterHost:     config.Host,
		KubeClient:     cli,
		HTTPClient:     httpc,
		DefaultTimeout: time.Minute,
	}

	return f, nil
}

func (f *Framework) CreateProxierOperator(namespace string, namespacesToWatch []string) error {
	_, err := CreateServiceAccount(f.KubeClient, namespace, "../../deploy/service_account.yaml")
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return errors.Wrap(err, "failed to create proxier operator service account")
	}

	if err := CreateClusterRole(f.KubeClient, "../../deploy/role.yaml"); err != nil && !apierrors.IsAlreadyExists(err) {
		return errors.Wrap(err, "failed to create proxier cluster role")
	}

	if _, err := CreateClusterRoleBinding(f.KubeClient, namespace, "../../deploy/role_binding.yaml"); err != nil && !apierrors.IsAlreadyExists(err) {
		return errors.Wrap(err, "failed to create proxier cluster role binding")
	}

	deployment, err := MakeDeployment("../../deploy/operator.yaml")
	if err != nil {
		return err
	}

	err = CreateDeployment(f.KubeClient, namespace, deployment)
	if err != nil {
		return err
	}

	opts := metav1.ListOptions{LabelSelector: fields.SelectorFromSet(fields.Set(deployment.Spec.Template.ObjectMeta.Labels)).String()}
	err = WaitForPodsReady(f.KubeClient, namespace, f.DefaultTimeout, 1, opts)
	if err != nil {
		return errors.Wrap(err, "failed to wait for prometheus operator to become ready")
	}

	err = k8sutil.WaitForCRDReady(func(opts metav1.ListOptions) (runtime.Object, error) {
		return f.MonClientV1.Prometheuses(v1.NamespaceAll).List(opts)
	})
	if err != nil {
		return errors.Wrap(err, "Proxier CRD not ready: %v\n")
	}

	return nil
}
