package v1beta1

const (
	ProxierKeyLabel = "maegus.proxier.key"
)

func GetProxierName(labels map[string]string) string {
	return labels[ProxierKeyLabel]
}
