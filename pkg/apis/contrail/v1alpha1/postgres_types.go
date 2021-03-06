package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Juniper/contrail-operator/pkg/certificates"
)

// +k8s:openapi-gen=true
type PostgresConfiguration struct {
	ListenPort                int          `json:"listenPort,omitempty"`
	RootPassSecretName        string       `json:"rootPassSecretName,omitempty"`
	ReplicationPassSecretName string       `json:"replicationPassSecretName,omitempty"`
	Containers                []*Container `json:"containers,omitempty"`
	Storage                   Storage      `json:"storage,omitempty"`
}

// PostgresSpec defines the desired state of Postgres
// +k8s:openapi-gen=true
type PostgresSpec struct {
	CommonConfiguration  PodConfiguration      `json:"commonConfiguration,omitempty"`
	ServiceConfiguration PostgresConfiguration `json:"serviceConfiguration"`
}

// PostgresStatus defines the observed state of Postgres
// +k8s:openapi-gen=true
type PostgresStatus struct {
	Status   `json:",inline"`
	Endpoint string `json:"endpoint,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Postgres is the Schema for the Postgress API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Replicas",type=integer,JSONPath=`.status.replicas`
// +kubebuilder:printcolumn:name="Ready_Replicas",type=integer,JSONPath=`.status.readyReplicas`
// +kubebuilder:printcolumn:name="Endpoint",type=string,JSONPath=`.status.endpoint`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// +kubebuilder:printcolumn:name="Active",type=boolean,JSONPath=`.status.active`
type Postgres struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PostgresSpec   `json:"spec,omitempty"`
	Status PostgresStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PostgresList contains a list of Postgres
type PostgresList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Postgres `json:"items"`
}

//PodsCertSubjects gets list of Postgres pods certificate subjets which can be passed to the certificate API
func (p *Postgres) PodsCertSubjects(podList *corev1.PodList, serviceIP string) []certificates.CertificateSubject {
	altIPs := PodAlternativeIPs{ServiceIP: serviceIP}
	return PodsCertSubjects(podList, p.Spec.CommonConfiguration.HostNetwork, altIPs)
}

// PostgresInstanceType is type unique name used for labels
const PostgresInstanceType = "postgres"

func init() {
	SchemeBuilder.Register(&Postgres{}, &PostgresList{})
}
