package v1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Rabbit is a top-level type
type Rabbit struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// This is where you can define
	// your own custom spec
	Spec RabbitSpec `json:"spec,omitempty"`
}

type RabbitSpec struct {
	Color string `json:"color,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RabbitList no client needed for list as it's been created in above
type RabbitList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `son:"metadata,omitempty"`

	Items []Rabbit `json:"items"`
}
