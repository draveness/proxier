package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ProxierSpec defines the desired state of Proxier
// +k8s:openapi-gen=true
type ProxierSpec struct {
	// +kubebuilder:validation:MinItems=1
	Servers []ServerSpec `json:"servers"`
}

// ServerSpec defines the target server of Proxier
type ServerSpec struct {
	Name string `json:"name"`

	Proportion float64 `json:"proportion"`

	TargetPort int32 `json:"targetPort"`

	// +kubebuilder:validation:MinItems=1
	Selector map[string]string `json:"selector,omitempty"`
}

// ProxierStatus defines the observed state of Proxier
// +k8s:openapi-gen=true
type ProxierStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Proxier is the Schema for the proxiers API
// +k8s:openapi-gen=true
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
