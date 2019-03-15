package proxier

import (
	"fmt"

	dravenessv1alpha1 "github.com/draveness/proxier/pkg/apis/draveness/v1alpha1"
)

// Server represents a nginx server
type Server struct {
	Name   string
	Weight int64
}

func newNginxConfigWithProxier(instance *dravenessv1alpha1.Proxier) string {
	servers := []Server{}
	for _, server := range instance.Spec.Servers {
		servers = append(servers, Server{
			Name:   fmt.Sprintf("%s-%s-server:%d", instance.Name, server.Name, server.TargetPort),
			Weight: int64(server.Proportion * 1000),
		})
	}

	conf := "events {\n"
	conf += "    worker_connections 1024;\n"
	conf += "}\n"
	conf += "http {\n"
	conf += "    upstream backend {\n"
	for _, server := range servers {
		conf += fmt.Sprintf("         server %s weight=%d;\n", server.Name, server.Weight)
	}
	conf += "    }\n"
	conf += "    server {\n"
	conf += "        location / {\n"
	conf += "            proxy_pass http://backend;\n"
	conf += "        }\n"
	conf += "    }\n"
	conf += "}\n"

	return conf
}
