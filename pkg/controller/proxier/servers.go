package proxier

import (
	"context"
	"fmt"
	"sync"

	maegusv1 "github.com/draveness/proxier/pkg/apis/maegus/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *ReconcileProxier) syncServers(instance *maegusv1.Proxier) error {
	var serviceList corev1.ServiceList
	if err := r.client.List(context.Background(), client.MatchingLabels(map[string]string{
		// TODO: use const for proxier name key in service
		"maegus.com/proxier-name": instance.Name,
	}), &serviceList); err != nil {
		return err
	}

	servicesToCreate, servicesToDelete := groupServers(instance, serviceList.Items)

	waitGroup := sync.WaitGroup{}
	waitGroup.Add(len(servicesToDelete))
	for i := range servicesToDelete {
		serviceToDelete := servicesToDelete[i]
		go func(service *corev1.Service) {
			defer waitGroup.Done()
			if err := r.client.Delete(context.Background(), service); err != nil {
				// TODO: handle delete service error
			}
		}(&serviceToDelete)
	}
	waitGroup.Wait()

	createErrCh := make(chan error, len(servicesToCreate))
	for i := range servicesToCreate {
		serviceToCreate := servicesToCreate[i]
		if err := controllerutil.SetControllerReference(instance, &serviceToCreate, r.scheme); err != nil {
			createErrCh <- err
			break
		}

		found := &corev1.Service{}
		err := r.client.Get(context.Background(), types.NamespacedName{Name: serviceToCreate.Name, Namespace: serviceToCreate.Namespace}, found)
		if err != nil && errors.IsNotFound(err) {
			err = r.client.Create(context.Background(), &serviceToCreate)
			if err != nil {
				createErrCh <- err
			}
			break

		} else if err != nil {
			createErrCh <- err
			break
		}

		found.Spec.Ports = serviceToCreate.Spec.Ports
		found.Spec.Selector = serviceToCreate.Spec.Selector

		err = r.client.Update(context.Background(), found)
		if err != nil {
			createErrCh <- err
			break
		}

	}

	select {
	case err := <-createErrCh:
		// all errors have been reported before and they're likely to be the same, so we'll only return the first one we hit.
		if err != nil {
			return err
		}
	default:
	}

	return nil
}

func groupServers(instance *maegusv1.Proxier, services []corev1.Service) ([]corev1.Service, []corev1.Service) {
	servicesToCreate := []corev1.Service{}

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

		serviceToCreate := corev1.Service{
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

		servicesToCreate = append(servicesToCreate, serviceToCreate)
	}

	servicesToDelete := []corev1.Service{}

	for i := range services {
		service := services[i]
		found := false

		for j := range servicesToCreate {
			serviceToCreate := servicesToCreate[j]
			if serviceToCreate.Name == service.Name &&
				serviceToCreate.Namespace == service.Namespace {
				found = true
				break
			}
		}

		if !found {
			servicesToDelete = append(servicesToDelete, service)
		}
	}

	return servicesToCreate, servicesToDelete
}
