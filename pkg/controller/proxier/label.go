package proxier

import (
	maegusv1 "github.com/draveness/proxier/pkg/apis/maegus/v1"
)

func newPodLabel(cr *maegusv1.Proxier) map[string]string {
	namespacedName := cr.Namespace + "." + cr.Name

	return map[string]string{
		"proxier": namespacedName,
	}
}
