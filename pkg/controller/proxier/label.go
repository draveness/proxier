package proxier

import (
	"fmt"

	maegusv1 "github.com/draveness/proxier/pkg/apis/maegus/v1beta1"
)

// NewPodLabels returns label used for pod.
func NewPodLabels(proxier *maegusv1.Proxier) map[string]string {
	return map[string]string{
		maegusv1.ProxierKeyLabel: fmt.Sprintf("%s", proxier.Name),
	}
}

// NewServiceLabels returns label used for service.
func NewServiceLabels(proxier *maegusv1.Proxier) map[string]string {
	return map[string]string{
		maegusv1.ProxierKeyLabel: fmt.Sprintf("%s", proxier.Name),
	}
}
