package proxier

import (
	"context"

	dravenessv1alpha1 "github.com/draveness/proxier/pkg/apis/draveness/v1alpha1"
	"github.com/draveness/proxier/pkg/controller/proxier/nginx"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *ReconcileProxier) syncDeployment(instance *dravenessv1alpha1.Proxier) error {
	// Sync ConfigMap for deployment
	newConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-proxy-configmap",
			Namespace: instance.Namespace,
		},
		Data: map[string]string{
			"nginx.conf": nginx.NewConfig(instance),
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

// newDeployment returns a busybox pod with the same name/namespace as the cr
func newDeployment(cr *dravenessv1alpha1.Proxier) *corev1.Pod {
	labels := newPodLabel(cr)

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-proxy",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "nginx",
					Image: "nginx:1.15.9",
					Ports: []corev1.ContainerPort{
						{
							ContainerPort: 80,
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      cr.Name + "-proxy-configmap",
							MountPath: "/etc/nginx",
							ReadOnly:  true,
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: cr.Name + "-proxy-configmap",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: cr.Name + "-proxy-configmap",
							},
						},
					},
				},
			},
		},
	}
}
