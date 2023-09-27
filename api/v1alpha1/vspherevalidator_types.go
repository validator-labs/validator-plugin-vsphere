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

	Auth                           VsphereAuth                          `json:"auth"`
	Datacenter                     string                               `json:"datacenter"`
	EntityPrivilegeValidationRules []EntityPrivilegeValidationRule      `json:"entityPrivilegeValidationRules"`
	RolePrivilegeValidationRules   []GenericRolePrivilegeValidationRule `json:"rolePrivilegeValidationRules"`
	TagValidationRules             []TagValidationRule                  `json:"tagValidationRules"`
	ComputeResourceRules           []ComputeResourceRule                `json:"computeResourceRules"`
}

type VsphereAuth struct {
	SecretName string `json:"secretName"`
}

type ComputeResourceRule struct {
	Name                         string                        `json:"name"`
	ClusterName                  string                        `json:"clusterName,omitempty"`
	Scope                        string                        `json:"scope"`
	EntityName                   string                        `json:"entityName"`
	NodepoolResourceRequirements []NodepoolResourceRequirement `json:"nodepoolResourceRequirements"`
}

type EntityPrivilegeValidationRule struct {
	Name        string   `json:"name"`
	ClusterName string   `json:"clusterName,omitempty"`
	EntityType  string   `json:"entityType"`
	EntityName  string   `json:"entityName"`
	Privileges  []string `json:"privileges"`
}

type GenericRolePrivilegeValidationRule struct {
	Name string `json:"name"`
}

type TagValidationRule struct {
	Name        string `json:"name"`
	ClusterName string `json:"clusterName,omitempty"`
	EntityType  string `json:"entityType"`
	EntityName  string `json:"entityName"`
	Tag         string `json:"tag"`
}

type NodepoolResourceRequirement struct {
	Name          string `json:"name"`
	NumberOfNodes int    `json:"numberOfNodes"`
	CPU           string `json:"cpu"`
	Memory        string `json:"memory"`
	DiskSpace     string `json:"diskSpace"`
}

type CloudAccountValidationRule struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	IsEnabled   bool     `json:"isEnabled"`
	Severity    string   `json:"severity"`
	RuleType    string   `json:"ruleType"`
	Expressions []string `json:"expressions"`
}

type DiskSpaceValidationRule struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	IsEnabled   bool     `json:"isEnabled"`
	Severity    string   `json:"severity"`
	RuleType    string   `json:"ruleType"`
	Expressions []string `json:"expressions"`
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
