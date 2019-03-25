package e2e

import (
	"testing"
)

func testCreateBasicProxier(t *testing.T) {
	t.Parallel()

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	ns := ctx.CreateNamespace(t, framework.KubeClient)
	ctx.SetupProxierRBAC(t, ns, framework.KubeClient)

	exampleProxier := framework.MakeBasicProxier(ns, "test", []string{"v1", "v2"}, []int32{100, 10})

	if _, err := framework.CreateProxierAndWaitUntilReady(ns, exampleProxier); err != nil {
		t.Fatal(err)
	}
}
