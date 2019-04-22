package e2e

import (
	"github.com/draveness/proxier/pkg/controller/proxier"
	"github.com/draveness/proxier/test/framework"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ProxierCreateSuite struct {
	suite.Suite
}

func (suite *ProxierCreateSuite) TestProxierCreateBackends() {
	ctx := f.NewTestCtx(suite.T())
	defer ctx.Cleanup(suite.T())

	namespace := ctx.CreateNamespace(suite.T(), f.KubeClient)
	ctx.SetupProxierRBAC(suite.T(), namespace, f.KubeClient)

	suite.T().Parallel()

	instance := framework.MakeBasicProxier(namespace, "test", []string{"v1", "v2"}, []int32{100, 10})

	_, err := f.CreateProxierAndWaitUntilReady(namespace, instance)

	suite.Nil(err, "create proxier error")

	svcList, err := f.KubeClient.CoreV1().Services(namespace).List(metav1.ListOptions{
		LabelSelector: "maegus.com/proxier-name=" + instance.Name,
	})

	suite.Nil(err, "list service error")
	suite.Equal(2, len(svcList.Items), "proxier should create backend services")
}

func (suite *ProxierCreateSuite) TestProxierCreateNginxDeployment() {
	ctx := f.NewTestCtx(suite.T())
	defer ctx.Cleanup(suite.T())

	namespace := ctx.CreateNamespace(suite.T(), f.KubeClient)
	ctx.SetupProxierRBAC(suite.T(), namespace, f.KubeClient)

	suite.T().Parallel()

	instance := framework.MakeBasicProxier(namespace, "test", []string{"v1", "v2"}, []int32{100, 10})

	_, err := f.CreateProxierAndWaitUntilReady(namespace, instance)
	suite.Nil(err, "create proxier error")

	deploymentName := proxier.NewDeploymentName(instance)
	deployment, err := f.KubeClient.AppsV1().Deployments(namespace).Get(deploymentName, metav1.GetOptions{})

	suite.Nil(err, "get deployment error")
	suite.Equal("nginx", deployment.Spec.Template.Spec.Containers[0].Name, "invalid nginx name")
	suite.Equal("nginx:1.15.9", deployment.Spec.Template.Spec.Containers[0].Image, "invalid nginx image")
}

func (suite *ProxierCreateSuite) TestProxierCreateService() {
	ctx := f.NewTestCtx(suite.T())
	defer ctx.Cleanup(suite.T())

	namespace := ctx.CreateNamespace(suite.T(), f.KubeClient)
	ctx.SetupProxierRBAC(suite.T(), namespace, f.KubeClient)

	suite.T().Parallel()

	instance := framework.MakeBasicProxier(namespace, "echo", []string{"v1", "v2"}, []int32{100, 10})

	_, err := f.CreateProxierAndWaitUntilReady(namespace, instance)
	suite.Nil(err, "create proxier error")

	proxierService, err := f.KubeClient.CoreV1().Services(namespace).Get(instance.Name, metav1.GetOptions{})
	suite.Nil(err, "get service error")
	suite.Equal(instance.Name, proxierService.Name, "proxier and service shoud have the same name")
}

func (suite *ProxierCreateSuite) TestUpdateProxierRemoveService() {
	ctx := f.NewTestCtx(suite.T())
	defer ctx.Cleanup(suite.T())

	namespace := ctx.CreateNamespace(suite.T(), f.KubeClient)
	ctx.SetupProxierRBAC(suite.T(), namespace, f.KubeClient)

	suite.T().Parallel()

	instance := framework.MakeBasicProxier(namespace, "echo", []string{"v1", "v2"}, []int32{100, 10})

	_, err := f.CreateProxierAndWaitUntilReady(namespace, instance)
	suite.Nil(err, "create proxier error")

	proxier, err := f.MaegusClientV1.Proxiers(instance.Namespace).Get(instance.Name, metav1.GetOptions{})
	suite.Nil(err)

	proxier.Spec = framework.MakeBasicProxier(namespace, "echo", []string{"v3"}, []int32{10}).Spec

	_, err = f.UpdateProxierAndWaitUntilReady(namespace, proxier)
	suite.Nil(err, "update proxier error")

	svcList, err := f.KubeClient.CoreV1().Services(namespace).List(metav1.ListOptions{
		LabelSelector: "maegus.com/proxier-name=" + instance.Name,
	})

	suite.Nil(err, "list service error")
	suite.Equal(1, len(svcList.Items), "proxier should remove useless backend services")
}
