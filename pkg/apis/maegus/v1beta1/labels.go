package v1beta1

const (
	// ProxierKeyLabel is the key of proxier name
	ProxierKeyLabel = "maegus.com/proxier-name"
)

// GetProxierName returns proxier name from labels.
func GetProxierName(labels map[string]string) string {
	return labels[ProxierKeyLabel]
}
