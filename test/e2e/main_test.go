package e2e

import (
	"flag"
	"log"
	"os"
	"testing"

	operatorFramework "github.com/draveness/proxier/test/framework"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/fields"
)

var (
	framework *operatorFramework.Framework
	opImage   *string
)

func TestMain(m *testing.M) {
	kubeconfig := flag.String(
		"kubeconfig",
		"",
		"kube config path, e.g. $HOME/.kube/config",
	)
	opImage = flag.String(
		"operator-image",
		"",
		"operator image, e.g. quay.io//prometheus-operator",
	)
	flag.Parse()

	var (
		err      error
		exitCode int
	)

	if framework, err = operatorFramework.New(*kubeconfig, *opImage); err != nil {
		log.Printf("failed to setup framework: %v\n", err)
		os.Exit(1)
	}

	exitCode = m.Run()

	os.Exit(exitCode)
}

func TestAllNS(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)

	ns := ctx.CreateNamespace(t, framework.KubeClient)

	err := framework.CreateProxierOperator(ns, nil)
	if err != nil {
		t.Fatal(err)
	}

	// t.Run blocks until the function passed as the second argument (f) returns or
	// calls t.Parallel to become a parallel test. Run reports whether f succeeded
	// (or at least did not fail before calling t.Parallel). As all tests in
	// testAllNS are parallel, the defered ctx.Cleanup above would be run before
	// all tests finished. Wrapping it in testAllNS fixes this.
	t.Run("x", testAllNS)

	// Check if Proxier Operator ever restarted.
	opts := metav1.ListOptions{LabelSelector: fields.SelectorFromSet(fields.Set(map[string]string{
		"apps.kubernetes.io/name": "proxier-operator",
	})).String()}

	pl, err := framework.KubeClient.CoreV1().Pods(ns).List(opts)
	if err != nil {
		t.Fatal(err)
	}
	if expected := 1; len(pl.Items) != expected {
		t.Fatalf("expected %v Proxier Operator pods, but got %v", expected, len(pl.Items))
	}
	restarts, err := framework.GetPodRestartCount(ns, pl.Items[0].GetName())
	if err != nil {
		t.Fatalf("failed to retrieve restart count of Proxier Operator pod: %v", err)
	}
	if len(restarts) != 1 {
		t.Fatalf("expected to have 1 container but got %d", len(restarts))
	}
	for _, restart := range restarts {
		if restart != 0 {
			t.Fatalf(
				"expected Proxier Operator to never restart during entire test execution but got %d restarts",
				restart,
			)
		}
	}
}

func testAllNS(t *testing.T) {
	testFuncs := map[string]func(t *testing.T){
		"CreateBasicProxier": testCreateBasicProxier,
	}

	for name, f := range testFuncs {
		t.Run(name, f)
	}
}
