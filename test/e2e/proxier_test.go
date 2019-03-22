package e2e

import (
	"fmt"
	"testing"

	operator "github.com/draveness/proxier/pkg/apis/maegus/v1"
	framework "github.com/operator-framework/operator-sdk/pkg/test"

	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/draveness/proxier/pkg/apis"
)

func TestProxier(t *testing.T) {
	proxierList := &operator.ProxierList{
		TypeMeta: metav1.TypesMeta{
			Kind:       "Proxier",
			APIVersion: "maegus.com/v1",
		},
	}

	err := framework.AddToFrameworkScheme(apis.AddToScheme, proxierList)
	if err != nil {
		t.Fatalf("Failed to add custom resource scheme to framework: %v", err)
	}

	t.Run("memcached-group", func(t *testing.T) {
		t.Run("Cluster", MemcachedCluster)
		t.Run("Cluster2", MemcachedCluster)
	})
}

func TestCreateBasicProxier(t *testing.T) error {
	t.Parallel()

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()

	err := ctx.InitializeClusterResources(&framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		t.Fatalf("Failed to initialize cluster resources: %v", err)
	}
	t.Log("Initialized cluster resources")
	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatal(err)
	}

	f := framework.Global
	err = e2eutil.WaitForOperatorDeployment(t, f.KubeClient, namespace, "memcached-operator", 1, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}

	namespace, err := ctx.GetNamespace()
	if err != nil {
		return fmt.Errorf("could not get namespace: %v", err)
	}

	exampleProxier := &operator.Proxier{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example",
			Namespace: namespace,
		},
		Spec: operator.ProxierSpec{
			Selector: map[string]string{
				"author": "draven",
				"app":    "blog",
			},
			Ports: []operator.ProxierPort{
				{
					Name:     "http",
					Protocol: operator.ProtocolTCP,
					Port:     80,
				},
			},
			Backends: []operator.BackendSpec{
				{
					Name:   "v1",
					Weight: 100,
					Selector: map[string]string{
						"version": "v1",
					},
				},
				{
					Name:   "v2",
					Weight: 10,
					Selector: map[string]string{
						"version": "v2",
					},
				},
			},
		},
	}

	err = f.Client.Create(goctx.TODO(), exampleProxier, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		return err
	}

	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "example-proxy", 1, retryInterval, timeout)
	if err != nil {
		return err
	}

}
