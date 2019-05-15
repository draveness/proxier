package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Protocol defines network protocols supported for things like container ports.
type Protocol string

const (
	// ProtocolTCP is the TCP protocol.
	ProtocolTCP Protocol = "TCP"
	// ProtocolUDP is the UDP protocol.
	ProtocolUDP Protocol = "UDP"
)

// ProxierPort contains information on proxier's port.
type ProxierPort struct {
	// The name of this port within the proxier. This must be a DNS_LABEL.
	// All ports within a ServiceSpec must have unique names. This maps to
	// the 'Name' field in EndpointPort objects.
	// Optional if only one ProxierPort is defined on this service.
	// +required
	Name string `json:"name,omitempty"`

	// The IP protocol for this port. Supports "TCP", "UDP".
	// Default is TCP.
	// +optional
	Protocol Protocol `json:"protocol,omitempty"`

	// The port that will be exposed by this proxier
	Port int32 `json:"port"`

	// +optional
	TargetPort intstr.IntOrString `json:"targetPort,omitempty"`
}

// ProxierSpec defines the desired state of Proxier
// +k8s:openapi-gen=true
type ProxierSpec struct {
	// +kubebuilder:validation:MinItems=1
	Backends []BackendSpec `json:"backends"`

	Selector map[string]string `json:"selector,omitempty"`

	Ports []ProxierPort `json:"ports"`
}

// BackendSpec defines the target backend of Proxier
type BackendSpec struct {
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// +kubebuilder:validation:Minimum=1
	Weight int32 `json:"weight"`

	Selector map[string]string `json:"selector,omitempty"`
}

// ProxierStatus defines the observed state of Proxier
// +k8s:openapi-gen=true
type ProxierStatus struct {
	// ActiveBackends stores the count of current active services, which are required by the current
	// proxier spec.
	// +optional
	ActiveBackends int32 `json:"activeBackends,omitempty"`

	// ObsoleteBackends stores the count of obsolete services, which should be removed in controller
	// +optional
	ObsoleteBackends int32 `json:"obsoleteBackends,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Proxier is the Schema for the proxiers API
// +genclient
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type Proxier struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProxierSpec   `json:"spec,omitempty"`
	Status ProxierStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ProxierList contains a list of Proxier
type ProxierList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Proxier `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Proxier{}, &ProxierList{})
}
