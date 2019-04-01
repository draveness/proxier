package e2e

import (
	"errors"
	"time"

	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/util/wait"
)

type ProxierCreateSuite struct {
	suite.Suite
	namespace string
}

func (suite *ProxierCreateSuite) SetupTest() {
	ctx := framework.NewTestCtx(suite.T())
	defer ctx.Cleanup(suite.T())
	ns := ctx.CreateNamespace(suite.T(), framework.KubeClient)
	ctx.SetupProxierRBAC(suite.T(), ns, framework.KubeClient)
}

func (suite *ProxierCreateSuite) testProxierCreateBackends() {
	suite.T().Parallel()

	exampleProxier := framework.MakeBasicProxier(suite.namespace, "test", []string{"v1", "v2"}, []int32{100, 10})

	if _, err := framework.CreateProxierAndWaitUntilReady(suite.namespace, exampleProxier); err != nil {
		suite.T().Fatal(err)
	}

	err := wait.Poll(5*time.Second, 30*time.Second, func() (bool, error) {
		svcList, err := framework.KubeClient.CoreV1().Services(suite.namespace).List(metav1.ListOptions{
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
		suite.T().Fatal(err)
	}
}
