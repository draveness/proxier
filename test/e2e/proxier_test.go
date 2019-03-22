package e2e

import (
	"testing"

	operator "github.com/draveness/proxier/pkg/apis/maegus/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func testCreateBasicProxier(t *testing.T) {
	t.Parallel()

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupProxierRBAC(t, ns, framework.KubeClient)

	exampleProxier := &operator.Proxier{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: ns,
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

	if _, err := framework.CreateProxierAndWaitUntilReady(ns, exampleProxier); err != nil {
		t.Fatal(err)
	}
}
