package proxier

import (
	"testing"

	maegusv1 "github.com/draveness/proxier/pkg/apis/maegus/v1beta1"

	"github.com/draveness/proxier/test/framework"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGroupServers(t *testing.T) {
	instance := framework.MakeBasicProxier("default", "group-server", []string{"v2", "v3"}, []int32{10, 20})

	existingServices := []corev1.Service{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "group-server-v1-backend",
				Namespace: "default",
				Labels: map[string]string{
					"maegus.com/proxier-name": instance.Name,
				},
			},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{
					"version": "v1",
					"app":     "group-server",
				},
				Type: corev1.ServiceTypeClusterIP,
				Ports: []corev1.ServicePort{
					{
						Name:     "http",
						Protocol: corev1.Protocol(maegusv1.ProtocolTCP),
						Port:     80,
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "group-server-v2-backend",
				Namespace: "default",
				Labels: map[string]string{
					"maegus.com/proxier-name": instance.Name,
				},
			},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{
					"version": "v2",
					"app":     "group-server",
				},
				Type: corev1.ServiceTypeClusterIP,
				Ports: []corev1.ServicePort{
					{
						Name:     "http",
						Protocol: corev1.Protocol(maegusv1.ProtocolTCP),
						Port:     80,
					},
				},
			},
		},
	}

	servicesToCreate, servicesToDelete := groupServers(instance, existingServices)

	assert.Equal(t, []corev1.Service{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "group-server-v1-backend",
				Namespace: "default",
				Labels: map[string]string{
					"maegus.com/proxier-name": instance.Name,
				},
			},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{
					"version": "v1",
					"app":     "group-server",
				},
				Type: corev1.ServiceTypeClusterIP,
				Ports: []corev1.ServicePort{
					{
						Name:     "http",
						Protocol: corev1.Protocol(maegusv1.ProtocolTCP),
						Port:     80,
					},
				},
			},
		},
	}, servicesToDelete)
	assert.Equal(t, []corev1.Service{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "group-server-v2-backend",
				Namespace: "default",
				Labels: map[string]string{
					"maegus.com/proxier-name": instance.Name,
				},
			},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{
					"version": "v2",
					"app":     "group-server",
				},
				Type: corev1.ServiceTypeClusterIP,
				Ports: []corev1.ServicePort{
					{
						Name:     "http",
						Protocol: corev1.Protocol(maegusv1.ProtocolTCP),
						Port:     80,
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "group-server-v3-backend",
				Namespace: "default",
				Labels: map[string]string{
					"maegus.com/proxier-name": instance.Name,
				},
			},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{
					"version": "v3",
					"app":     "group-server",
				},
				Type: corev1.ServiceTypeClusterIP,
				Ports: []corev1.ServicePort{
					{
						Name:     "http",
						Protocol: corev1.Protocol(maegusv1.ProtocolTCP),
						Port:     80,
					},
				},
			},
		},
	}, servicesToCreate)
}
