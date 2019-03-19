package proxier

import (
	"context"
	"fmt"
	"sync"

	dravenessv1alpha1 "github.com/draveness/proxier/pkg/apis/draveness/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *ReconcileProxier) syncServers(instance *dravenessv1alpha1.Proxier) error {
	serversCount := len(instance.Spec.Servers)

	services := []corev1.Service{}
	for _, server := range instance.Spec.Servers {
		service := corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-%s-server", instance.Name, server.Name),
				Namespace: instance.Namespace,
			},
			Spec: corev1.ServiceSpec{
				Selector: server.Selector,
				Type:     corev1.ServiceTypeClusterIP,
				Ports: []corev1.ServicePort{
					corev1.ServicePort{
						Name:     "proxy",
						Port:     server.TargetPort,
						Protocol: corev1.ProtocolTCP,
					},
				},
			},
		}

		services = append(services, service)
	}

	errCh := make(chan error, serversCount)

	var wg sync.WaitGroup
	wg.Add(serversCount)

	for i := range services {
		service := services[i]
		go func(service *corev1.Service) {
			defer wg.Done()

			if err := controllerutil.SetControllerReference(instance, service, r.scheme); err != nil {
				errCh <- err
				return
			}

			found := &corev1.Service{}
			err := r.client.Get(context.TODO(), types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, found)
			if err != nil && errors.IsNotFound(err) {
				err = r.client.Create(context.TODO(), service)
				if err != nil {
					errCh <- err
				}

			} else if err != nil {
				errCh <- err
			}

			found.Spec.Ports = service.Spec.Ports
			found.Spec.Selector = service.Spec.Selector

			err = r.client.Update(context.TODO(), found)
			if err != nil {
				errCh <- err
			}

		}(&service)
	}
	wg.Wait()

	select {
	case err := <-errCh:
		// all errors have been reported before and they're likely to be the same, so we'll only return the first one we hit.
		if err != nil {
			return err
		}
	default:
	}

	return nil
}
