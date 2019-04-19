package e2e

import (
	"flag"
	"log"
	"os"
	"testing"

	"github.com/draveness/proxier/test/framework"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/fields"
)

var (
	f             *framework.Framework
	operatorImage *string
)

func TestMain(m *testing.M) {
	kubeconfig := flag.String(
		"kubeconfig",
		"",
		"kube config path, e.g. $HOME/.kube/config",
	)
	operatorImage = flag.String(
		"operator-image",
		"",
		"operator image, e.g. draveness/proxier:v1.0.0",
	)
	flag.Parse()

	var (
		err      error
		exitCode int
	)

	if f, err = framework.New(*kubeconfig, *operatorImage); err != nil {
		log.Printf("failed to setup framework: %v\n", err)
		os.Exit(1)
	}

	exitCode = m.Run()

	os.Exit(exitCode)
}

func TestAllNS(t *testing.T) {
	ctx := f.NewTestCtx(t)
	defer ctx.Cleanup(t)

	ns := ctx.CreateNamespace(t, f.KubeClient)

	err := f.CreateProxierOperator(ns, *operatorImage, nil)
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
		"name": "proxier-operator",
	})).String()}

	pl, err := f.KubeClient.CoreV1().Pods(ns).List(opts)
	if err != nil {
		t.Fatal(err)
	}
	if expected := 1; len(pl.Items) != expected {
		t.Fatalf("expected %v Proxier Operator pods, but got %v", expected, len(pl.Items))
	}
	restarts, err := f.GetPodRestartCount(ns, pl.Items[0].GetName())
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
	suite.Run(t, new(ProxierCreateSuite))
}
