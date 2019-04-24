package proxier

import (
	maegusv1 "github.com/draveness/proxier/pkg/apis/maegus/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/apis/autoscaling"
)

func (r *ReconcileProxier) syncHPA(instance *maegusv1.Proxier) error {
}

// NewHPA returns the horizontal pod autoscaler which is responsible for the autoscaling
// of deployment's replicas.
func NewHPA(instance *maegusv1.Proxier) *autoscaling.HorizontalPodAutoscaler {
	return &autoscaling.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-hpa",
			Namespace: instance.Namespace,
		},
		Spec: autoscaling.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscaling.CrossVersionObjectReference{
				APIVersion: "",
				Kind:       "Deployment",
				Name:       "",
			},
			MaxReplicas: 8,
			Metrics: []autoscaling.MetricSpec{
				{},
			},
		},
	}
}
