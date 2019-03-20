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
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *ReconcileProxier) syncServers(instance *dravenessv1alpha1.Proxier) error {
	backendsCount := len(instance.Spec.Backends)

	proxierSelector := instance.Spec.Selector

	proxierPorts := []corev1.ServicePort{}
	for _, port := range instance.Spec.Ports {
		proxierPorts = append(proxierPorts, corev1.ServicePort{
			Name:       port.Name,
			Protocol:   corev1.Protocol(port.Protocol),
			Port:       port.Port,
			TargetPort: port.TargetPort,
		})
	}

	services := []corev1.Service{}
	for _, backend := range instance.Spec.Backends {
		backendSelector := proxierSelector
		for key, value := range backend.Selector {
			backendSelector[key] = value
		}

		mergedPorts := proxierPorts
		for i, proxierPort := range proxierPorts {
			for _, backendPort := range backend.Ports {
				if backendPort.Name == proxierPort.Name {
					mergedPorts[i].TargetPort = intstr.FromInt(int(backendPort.TargetPort))
				}
			}
		}

		service := corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-%s-backend", instance.Name, backend.Name),
				Namespace: instance.Namespace,
			},
			Spec: corev1.ServiceSpec{
				Selector: backendSelector,
				Type:     corev1.ServiceTypeClusterIP,
				Ports:    mergedPorts,
			},
		}

		services = append(services, service)
	}

	errCh := make(chan error, backendsCount)

	var wg sync.WaitGroup
	wg.Add(backendsCount)

	for i := range services {
		service := services[i]
		go func(service corev1.Service) {
			defer wg.Done()

			if err := controllerutil.SetControllerReference(instance, &service, r.scheme); err != nil {
				errCh <- err
				return
			}

			found := &corev1.Service{}
			err := r.client.Get(context.TODO(), types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, found)
			if err != nil && errors.IsNotFound(err) {
				err = r.client.Create(context.TODO(), &service)
				if err != nil {
					errCh <- err
				}
				return

			} else if err != nil {
				errCh <- err
				return
			}

			found.Spec.Ports = service.Spec.Ports
			found.Spec.Selector = service.Spec.Selector

			err = r.client.Update(context.TODO(), found)
			if err != nil {
				errCh <- err
			}

		}(service)
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

func (r *ReconcileProxier) syncDeployment(instance *dravenessv1alpha1.Proxier) error {
	// Sync ConfigMap for deployment
	newConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-proxy-configmap",
			Namespace: instance.Namespace,
		},
		Data: map[string]string{
			"nginx.conf": newNginxConfigWithProxier(instance),
		},
	}

	if err := controllerutil.SetControllerReference(instance, newConfigMap, r.scheme); err != nil {
		return err
	}

	foundConfigMap := &corev1.ConfigMap{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: newConfigMap.Name, Namespace: newConfigMap.Namespace}, foundConfigMap)
	if err != nil && errors.IsNotFound(err) {
		err = r.client.Create(context.TODO(), newConfigMap)
		if err != nil {
			return err
		}

		// ConfigMap created successfully - don't requeue
		return nil
	} else if err != nil {
		return err
	}

	foundConfigMap.Data = newConfigMap.Data

	err = r.client.Update(context.TODO(), foundConfigMap)
	if err != nil {
		return err
	}

	pod := newDeployment(instance)

	if err := controllerutil.SetControllerReference(instance, pod, r.scheme); err != nil {
		return err
	}

	// Check if this Pod already exists
	foundPod := &corev1.Pod{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, foundPod)
	if err != nil && errors.IsNotFound(err) {
		err = r.client.Create(context.TODO(), pod)
		if err != nil {
			return err
		}

		// Pod created successfully - don't requeue
		return nil
	} else if err != nil {
		return err
	}

	return nil
}
