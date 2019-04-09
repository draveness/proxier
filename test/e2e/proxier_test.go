package e2e

import (
	"github.com/draveness/proxier/pkg/controller/proxier"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	_, err := framework.CreateProxierAndWaitUntilReady(suite.namespace, exampleProxier)

	assert.Nil(suite.T(), err, "create proxier error")

	svcList, err := framework.KubeClient.CoreV1().Services(suite.namespace).List(metav1.ListOptions{
		LabelSelector: "maegus.com/proxier-name=" + exampleProxier.Name,
	})

	assert.Nil(suite.T(), err, "list service error")
	assert.Equal(suite.T(), 2, len(svcList.Items), "proxier should create backend services")

	deploymentName := proxier.NewDeploymentName(exampleProxier)
	deployment, err := framework.KubeClient.AppsV1().Deployments(suite.namespace).Get(deploymentName, metav1.GetOptions{})

	assert.Nil(suite.T(), err, "get deployment error")
	assert.Equal(suite.T(), "nginx:1.15.9", deployment.Spec.Template.Spec.Containers[0].Image, "invalid nginx image")
}
