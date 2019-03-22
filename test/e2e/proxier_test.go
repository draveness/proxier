package e2e

import (
	"testing"

	operator "github.com/draveness/proxier/pkg/apis/maegus/v1"
	"github.com/draveness/proxier/pkg/test/framework"

	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func testCreateBasicProxier(t *testing.T) error {
	t.Parallel()

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()
	ns := ctx.CreateNamespace(t, framework.KubeClient)

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
