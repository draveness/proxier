package e2e

import (
	"errors"
	"fmt"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
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

	err := wait.Poll(5*time.Second, 30*time.Second, func() (bool, error) {
		svcList1, err := framework.KubeClient.CoreV1().Services(ns).List(metav1.ListOptions{})
		fmt.Println(svcList1, err)
		svcList, err := framework.KubeClient.CoreV1().Services(ns).List(metav1.ListOptions{
			LabelSelector: "maegus.com/proxier-name=" + exampleProxier.Name,
		})
		if err != nil {
			return false, err
		}

		if len(svcList.Items) != 2 {
			return false, errors.New("proxier should create backend services")
		}

		return true, nil
	})

	if err != nil {
		t.Fatal(err)
	}
}
