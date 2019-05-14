package proxier

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	maegusv1 "github.com/draveness/proxier/pkg/apis/maegus/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *ReconcileProxier) syncProxierStatus(instance *maegusv1.Proxier) error {
	// TODO: calculate status in syncProxierStatus
	var newStatus maegusv1.ProxierStatus

	if reflect.DeepEqual(instance.Status, newStatus) {
		return nil
	}

	newProxier := instance
	newProxier.Status = newStatus
	return r.client.Status().Update(context.Background(), newProxier)
}

func (r *ReconcileProxier) syncServers(instance *maegusv1.Proxier) error {
	var serviceList corev1.ServiceList
	if err := r.client.List(context.Background(), client.MatchingLabels(NewServiceLabels(instance)), &serviceList); err != nil {
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
				// TODO: handle delete service error in sync servers
			}
		}(&serviceToDelete)
	}
	waitGroup.Wait()

	createErrCh := make(chan error, len(servicesToCreate))
	waitGroup.Add(len(servicesToCreate))
	for i := range servicesToCreate {
		serviceToCreate := servicesToCreate[i]
		go func(service *corev1.Service) {
			defer waitGroup.Done()
			if err := controllerutil.SetControllerReference(instance, service, r.scheme); err != nil {
				createErrCh <- err
				return
			}

			found := &corev1.Service{}
			err := r.client.Get(context.Background(), types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, found)
			if err != nil && errors.IsNotFound(err) {
				err = r.client.Create(context.Background(), service)
				if err != nil {
					createErrCh <- err
				}
				return

			} else if err != nil {
				createErrCh <- err
				return
			}

			found.Spec.Ports = service.Spec.Ports
			found.Spec.Selector = service.Spec.Selector

			err = r.client.Update(context.Background(), found)
			if err != nil {
				createErrCh <- err
				return
			}
		}(&serviceToCreate)
	}
	waitGroup.Wait()

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
				Labels:    NewServiceLabels(instance),
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
