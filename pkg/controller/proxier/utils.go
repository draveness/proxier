package proxier

import (
	dravenessv1alpha1 "github.com/draveness/proxier/pkg/apis/draveness/v1alpha1"
)

func newPodLabel(cr *dravenessv1alpha1.Proxier) map[string]string {
	namespacedName := cr.Namespace + "." + cr.Name

	return map[string]string{
		"proxier": namespacedName,
	}
}
