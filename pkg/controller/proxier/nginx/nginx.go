package nginx

import (
	"fmt"

	dravenessv1alpha1 "github.com/draveness/proxier/pkg/apis/draveness/v1alpha1"
)

type server struct {
	name     string
	protocol string
	port     int32
}

type upstream struct {
	name     string
	backends []backend
}

type backend struct {
	name   string
	weight int32
}

func NewConfig(instance *dravenessv1alpha1.Proxier) string {
	servers := []server{}
	for _, port := range instance.Spec.Ports {
		server := server{
			name:     port.Name,
			protocol: string(port.Protocol),
			port:     port.Port,
			upstream: upstream,
		}
		servers = append(servers, server)
	}

	backends := []backend{}
	for _, be := range instance.Spec.Backends {
		backend := backend{
			name:   fmt.Sprintf("%s-%s-backendd", instance.Name, be.Name),
			weight: be.Weight,
		}
		backends = append(backends, backend)
	}

}
