package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ClusterJobSpec defines the desired state of ClusterJob
// +k8s:openapi-gen=true
type ClusterJobSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	// https://github.com/kubernetes/kube-openapi/issues/175
	// +listType=map
	JobImages map[string]string `json:"jobImages"`
}

// ClusterJobStatus defines the observed state of ClusterJob
// +k8s:openapi-gen=true
type ClusterJobStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	// +k8s:openapi-gen=true
	AllDone        bool                 `json:"allDone"`
	AllSucceeded   bool                 `json:"allSucceeded"`
	TotalStarted   uint8                `json:"totalStarted"`
	TotalSucceeded uint8                `json:"totalSucceeded"`
	TotalFailed    uint8                `json:"totalFailed"`
	JobStatuses    map[string]JobStatus `json:"JobStatuses"`
}

type JobStatus uint8

// kubectl explain jobs.status
const (
	ACTIVE    JobStatus = iota
	SUCCEEDED           = 1
	FAILED              = 2
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterJob is the Schema for the clusterjobs API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=clusterjobs,scope=Namespaced
type ClusterJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterJobSpec   `json:"spec,omitempty"`
	Status ClusterJobStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterJobList contains a list of ClusterJob
type ClusterJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterJob{}, &ClusterJobList{})
}
