package proxier

import (
	maegusv1 "github.com/draveness/proxier/pkg/apis/maegus/v1"
	"k8s.io/kubernetes/pkg/controller"
)

func newPodLabel(proxier *maegusv1.Proxier) (map[string]string, error) {
	key, err := controller.KeyFunc(proxier)
	if err != nil {
		return nil, err
	}

	return map[string]string{
		maegusv1.ProxierNameLabel: key,
	}, nil
}
