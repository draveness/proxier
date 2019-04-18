package proxier

import (
	"fmt"

	maegusv1 "github.com/draveness/proxier/pkg/apis/maegus/v1"
)

func newPodLabel(proxier *maegusv1.Proxier) (map[string]string, error) {
	return map[string]string{
		maegusv1.ProxierKeyLabel: fmt.Sprintf("%s.%s", proxier.Name, proxier.Namespace),
	}, nil
}
