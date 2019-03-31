package framework

import (
	"net/http"
	"testing"
	"time"

	"github.com/coreos/prometheus-operator/pkg/k8sutil"
	maegusclient "github.com/draveness/proxier/pkg/client/versioned/typed/maegus/v1"
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
	MaegusClientV1 maegusclient.MaegusV1Interface
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

	mClientV1, err := maegusclient.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "creating v1 maegus client failed")
	}

	f := &Framework{
		MasterHost:     config.Host,
		MaegusClientV1: mClientV1,
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

	if err := CreateClusterRole(f.KubeClient, "../../deploy/cluster_role.yaml"); err != nil && !apierrors.IsAlreadyExists(err) {
		return errors.Wrap(err, "failed to create proxier cluster role")
	}

	if _, err := CreateClusterRoleBinding(f.KubeClient, namespace, "../../deploy/cluster_role_binding.yaml"); err != nil && !apierrors.IsAlreadyExists(err) {
		return errors.Wrap(err, "failed to create prometheus cluster role binding")
	}

	// TODO: update image
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
		return f.MaegusClientV1.Proxiers(v1.NamespaceAll).List(opts)
	})
	if err != nil {
		return errors.Wrap(err, "Proxier CRD not ready: %v\n")
	}

	return nil
}

func (ctx *TestCtx) SetupProxierRBAC(t *testing.T, ns string, kubeClient kubernetes.Interface) {
	if err := CreateClusterRole(kubeClient, "../../deploy/cluster_role.yaml"); err != nil && !apierrors.IsAlreadyExists(err) {
		t.Fatalf("failed to create proxier cluster role: %v", err)
	}
	if finalizerFn, err := CreateServiceAccount(kubeClient, ns, "../../deploy/service_account.yaml"); err != nil {
		t.Fatal(errors.Wrap(err, "failed to create proxier service account"))
	} else {
		ctx.AddFinalizerFn(finalizerFn)
	}

	if finalizerFn, err := CreateRoleBinding(kubeClient, ns, "../../deploy/cluster_role_binding.yaml"); err != nil {
		t.Fatal(errors.Wrap(err, "failed to create proxier role binding"))
	} else {
		ctx.AddFinalizerFn(finalizerFn)
	}
}
