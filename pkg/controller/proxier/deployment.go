package proxier

import (
	"context"

	dravenessv1alpha1 "github.com/draveness/proxier/pkg/apis/draveness/v1alpha1"
	"github.com/draveness/proxier/pkg/controller/proxier/nginx"
	appsv1 "k8s.io/api/apps/v1"
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

	deployment := newDeployment(instance)

	annotations := map[string]string{}
	annotations["draveness.me/proxier-config-hash"] = computeHash(newConfigMap)
	deployment.Spec.Template.ObjectMeta.Annotations = annotations

	if err := controllerutil.SetControllerReference(instance, deployment, r.scheme); err != nil {
		return err
	}

	// Check if this Deployment already exists
	foundDeployment := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace}, foundDeployment)
	if err != nil && errors.IsNotFound(err) {
		err = r.client.Create(context.TODO(), deployment)
		if err != nil {
			return err
		}

		// Deployment created successfully - don't requeue
		return nil
	} else if err != nil {
		return err
	}

	foundDeployment.Spec.Template = deployment.Spec.Template

	err = r.client.Update(context.TODO(), foundDeployment)
	if err != nil {
		return err
	}

	return nil
}

// newDeployment returns a busybox pod with the same name/namespace as the cr
func newDeployment(cr *dravenessv1alpha1.Proxier) *appsv1.Deployment {
	labels := newPodLabel(cr)

	replicas := int32(1)

	deployment := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-proxy",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
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
			},
		},
	}

	return &deployment
}
