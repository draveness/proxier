package proxier

import (
	"context"
	"fmt"

	maegusv1 "github.com/draveness/proxier/pkg/apis/maegus/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *ReconcileProxier) syncServers(instance *maegusv1.Proxier) error {
	servicesToCreate, _ := groupServers(instance, []*corev1.Service{})

	backendsCount := len(instance.Spec.Backends)
	errCh := make(chan error, backendsCount)
	for i := range servicesToCreate {
		service := servicesToCreate[i]
		if err := controllerutil.SetControllerReference(instance, service, r.scheme); err != nil {
			errCh <- err
			break
		}

		found := &corev1.Service{}
		err := r.client.Get(context.Background(), types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, found)
		if err != nil && errors.IsNotFound(err) {
			err = r.client.Create(context.Background(), service)
			if err != nil {
				errCh <- err
			}
			break

		} else if err != nil {
			errCh <- err
			break
		}

		found.Spec.Ports = service.Spec.Ports
		found.Spec.Selector = service.Spec.Selector

		err = r.client.Update(context.Background(), found)
		if err != nil {
			errCh <- err
			break
		}

	}

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

func groupServers(instance *maegusv1.Proxier, services []*corev1.Service) ([]*corev1.Service, []*corev1.Service) {
	// TODO: delete useless servers
	servicesToDelete := []*corev1.Service{}

	servicesToCreate := []*corev1.Service{}

	proxierPorts := []corev1.ServicePort{}
	for _, port := range instance.Spec.Ports {
		proxierPorts = append(proxierPorts, corev1.ServicePort{
			Name:       port.Name,
			Protocol:   corev1.Protocol(port.Protocol),
			Port:       port.Port,
			TargetPort: port.TargetPort,
		})
	}

	for _, backend := range instance.Spec.Backends {
		backendSelector := map[string]string{}
		for key, value := range instance.Spec.Selector {
			backendSelector[key] = value
		}
		for key, value := range backend.Selector {
			backendSelector[key] = value
		}

		service := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-%s-backend", instance.Name, backend.Name),
				Namespace: instance.Namespace,
				Labels: map[string]string{
					"maegus.com/proxier-name": instance.Name,
				},
			},
			Spec: corev1.ServiceSpec{
				Selector: backendSelector,
				Type:     corev1.ServiceTypeClusterIP,
				Ports:    proxierPorts,
			},
		}

		servicesToCreate = append(servicesToCreate, service)
	}

	return servicesToCreate, servicesToDelete
}
