package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// VsphereValidatorSpec defines the desired state of VsphereValidator
type VsphereValidatorSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Auth                           VsphereAuth                          `json:"auth" yaml:"auth"`
	Datacenter                     string                               `json:"datacenter" yaml:"datacenter"`
	EntityPrivilegeValidationRules []EntityPrivilegeValidationRule      `json:"entityPrivilegeValidationRules" yaml:"entityPrivilegeValidationRules"`
	RolePrivilegeValidationRules   []GenericRolePrivilegeValidationRule `json:"rolePrivilegeValidationRules" yaml:"rolePrivilegeValidationRules"`
	TagValidationRules             []TagValidationRule                  `json:"tagValidationRules" yaml:"tagValidationRules"`
	ComputeResourceRules           []ComputeResourceRule                `json:"computeResourceRules" yaml:"computeResourceRules"`
}

type VsphereAuth struct {
	SecretName string `json:"secretName"`
}

type ComputeResourceRule struct {
	Name                         string                        `json:"name" yaml:"name"`
	ClusterName                  string                        `json:"clusterName,omitempty" yaml:"clusterName"`
	Scope                        string                        `json:"scope" yaml:"scope"`
	EntityName                   string                        `json:"entityName" yaml:"entityName"`
	NodepoolResourceRequirements []NodepoolResourceRequirement `json:"nodepoolResourceRequirements" yaml:"nodepoolResourceRequirements"`
}

type EntityPrivilegeValidationRule struct {
	Name        string   `json:"name" yaml:"name"`
	Username    string   `json:"username" yaml:"username"`
	ClusterName string   `json:"clusterName,omitempty" yaml:"clusterName"`
	EntityType  string   `json:"entityType" yaml:"entityType"`
	EntityName  string   `json:"entityName" yaml:"entityName"`
	Privileges  []string `json:"privileges" yaml:"privileges"`
}

type GenericRolePrivilegeValidationRule struct {
	Username   string   `json:"username" yaml:"username"`
	Privileges []string `json:"privileges" yaml:"privileges"`
}

type TagValidationRule struct {
	Name        string `json:"name" yaml:"name"`
	ClusterName string `json:"clusterName,omitempty" yaml:"clusterName"`
	EntityType  string `json:"entityType" yaml:"entityType"`
	EntityName  string `json:"entityName" yaml:"entityName"`
	Tag         string `json:"tag" yaml:"tag"`
}

type NodepoolResourceRequirement struct {
	Name          string `json:"name" yaml:"name"`
	NumberOfNodes int    `json:"numberOfNodes" yaml:"numberOfNodes"`
	CPU           string `json:"cpu" yaml:"cpu"`
	Memory        string `json:"memory" yaml:"memory"`
	DiskSpace     string `json:"diskSpace" yaml:"diskSpace"`
}

// VsphereValidatorStatus defines the observed state of VsphereValidator
type VsphereValidatorStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// VsphereValidator is the Schema for the vspherevalidators API
type VsphereValidator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VsphereValidatorSpec   `json:"spec,omitempty"`
	Status VsphereValidatorStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// VsphereValidatorList contains a list of VsphereValidator
type VsphereValidatorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VsphereValidator `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VsphereValidator{}, &VsphereValidatorList{})
}
