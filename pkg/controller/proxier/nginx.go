package proxier

import (
	dravenessv1alpha1 "github.com/draveness/proxier/pkg/apis/draveness/v1alpha1"
)

// Server represents a nginx backend
type Server struct {
	Name   string
	Weight int64
}

func newNginxConfigWithProxier(instance *dravenessv1alpha1.Proxier) string {
	// servers := []Server{}
	// for _, backend := range instance.Spec.Backends {
	// 	servers = append(servers, Server{
	// 		Name:   fmt.Sprintf("%s-%s-backend:%d", instance.Name, backend.Name, backend.TargetPort),
	// 		Weight: int64(backend.Proportion * 1000),
	// 	})
	// }

	// conf := "events {\n"
	// conf += "    worker_connections 1024;\n"
	// conf += "}\n"
	// conf += "http {\n"
	// conf += "    upstream backend {\n"
	// for _, backend := range servers {
	// 	conf += fmt.Sprintf("         backend %s weight=%d;\n", backend.Name, backend.Weight)
	// }
	// conf += "    }\n"
	// conf += "    backend {\n"
	// conf += "        location / {\n"
	// conf += "            proxy_pass http://backend;\n"
	// conf += "        }\n"
	// conf += "    }\n"
	// conf += "}\n"

	conf := ""

	return conf
}
