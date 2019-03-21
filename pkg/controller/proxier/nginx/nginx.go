package nginx

import (
	"fmt"
	"strings"

	dravenessv1alpha1 "github.com/draveness/proxier/pkg/apis/draveness/v1alpha1"
)

type server struct {
	name     string
	protocol string
	port     int32
	upstream string
}

func (s server) conf() string {
	var protocol string
	if string(s.protocol) == "udp" {
		protocol = "udp"
	}
	return fmt.Sprintf(`
server {
    listen %d %s;
    proxy_pass %s;
}
`, s.port, protocol, s.upstream)
}

type upstream struct {
	name     string
	backends []backend
	port     int32
}

func (up upstream) conf() string {
	backendStrs := ""
	for _, backend := range up.backends {
		backendStrs += fmt.Sprintf("    server %s:%d weight=%d;\n", backend.name, up.port, backend.weight)
	}

	return fmt.Sprintf(`
upstream %s {
%s
}
`, up.name, backendStrs)
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
			protocol: strings.ToLower(string(port.Protocol)),
			port:     port.Port,
			upstream: fmt.Sprintf("upstream_%s", port.Name),
		}
		servers = append(servers, server)
	}

	backends := []backend{}
	for _, be := range instance.Spec.Backends {
		backend := backend{
			name:   fmt.Sprintf("%s-%s-backend", instance.Name, be.Name),
			weight: be.Weight,
		}
		backends = append(backends, backend)
	}

	upstreams := []upstream{}
	for _, server := range servers {
		upstreams = append(upstreams, upstream{
			name:     server.upstream,
			port:     server.port,
			backends: backends,
		})
	}

	conf := ""
	conf += "events {\n"
	conf += "    worker_connections 1024;\n"
	conf += "}\n"
	conf += "stream {\n"

	for _, server := range servers {
		conf += server.conf()
	}

	for _, upstream := range upstreams {
		conf += upstream.conf()
	}

	conf += "}\n"

	return conf
}
