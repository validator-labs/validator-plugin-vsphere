package v1alpha1

import (
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/validator-labs/validator/pkg/plugins"
	"github.com/validator-labs/validator/pkg/validationrule"

	"github.com/validator-labs/validator-plugin-vsphere/pkg/constants"
	"github.com/validator-labs/validator-plugin-vsphere/pkg/vsphere"
)

// VsphereValidatorSpec defines the desired state of VsphereValidator
type VsphereValidatorSpec struct {
	Auth                     VsphereAuth               `json:"auth" yaml:"auth"`
	Datacenter               string                    `json:"datacenter" yaml:"datacenter"`
	PrivilegeValidationRules []PrivilegeValidationRule `json:"privilegeValidationRules,omitempty" yaml:"privilegeValidationRules,omitempty"`
	TagValidationRules       []TagValidationRule       `json:"tagValidationRules,omitempty" yaml:"tagValidationRules,omitempty"`
	ComputeResourceRules     []ComputeResourceRule     `json:"computeResourceRules,omitempty" yaml:"computeResourceRules,omitempty"`
	NTPValidationRules       []NTPValidationRule       `json:"ntpValidationRules,omitempty" yaml:"ntpValidationRules,omitempty"`
}

var _ plugins.PluginSpec = (*VsphereValidatorSpec)(nil)

// PluginCode returns the vSphere validator's plugin code.
func (s VsphereValidatorSpec) PluginCode() string {
	return constants.PluginCode
}

// ResultCount returns the number of validation results expected for an VsphereValidatorSpec.
func (s VsphereValidatorSpec) ResultCount() int {
	return len(s.PrivilegeValidationRules) + len(s.ComputeResourceRules) +
		len(s.TagValidationRules) + len(s.NTPValidationRules)
}

// VsphereAuth defines authentication configuration for an VsphereValidator.
type VsphereAuth struct {
	// SecretName is the name of the secret containing the vSphere credentials
	SecretName string `json:"secretName,omitempty" yaml:"secretName,omitempty"`

	// Account is the vSphere account to use for authentication
	Account *vsphere.Account `json:"account,omitempty" yaml:"account,omitempty"`
}

// NTPValidationRule defines the NTP validation rule
type NTPValidationRule struct {
	validationrule.ManuallyNamed `json:",inline" yaml:",omitempty"`

	// RuleName is the name of the NTP validation rule
	RuleName string `json:"name" yaml:"name"`

	// ClusterName is required when the vCenter Host(s) reside beneath a Cluster in the vCenter object hierarchy
	ClusterName string `json:"clusterName,omitempty" yaml:"clusterName,omitempty"`

	// Hosts is the list of vCenter Hosts to validate NTP configuration
	Hosts []string `json:"hosts" yaml:"hosts"`
}

var _ validationrule.Interface = (*NTPValidationRule)(nil)

// Name returns the name of the NTPValidationRule.
func (r NTPValidationRule) Name() string {
	return r.RuleName
}

// SetName sets the name of the NTPValidationRule.
func (r *NTPValidationRule) SetName(name string) {
	r.RuleName = name
}

// ComputeResourceRule defines the compute resource validation rule
type ComputeResourceRule struct {
	validationrule.ManuallyNamed `json:",inline" yaml:",omitempty"`

	// RuleName is the name of the compute resource validation rule
	RuleName string `json:"name" yaml:"name"`

	// ClusterName is required when the vCenter Entity resides beneath a Cluster in the vCenter object hierarchy
	ClusterName string `json:"clusterName,omitempty" yaml:"clusterName"`

	// Scope is the scope of the compute resource validation rule
	Scope string `json:"scope" yaml:"scope"`

	// EntityName is the name of the entity to validate
	EntityName string `json:"entityName" yaml:"entityName"`

	// NodepoolResourceRequirements is the list of nodepool resource requirements
	NodepoolResourceRequirements []NodepoolResourceRequirement `json:"nodepoolResourceRequirements" yaml:"nodepoolResourceRequirements"`
}

var _ validationrule.Interface = (*ComputeResourceRule)(nil)

// Name returns the name of the ComputeResourceRule.
func (r ComputeResourceRule) Name() string {
	return r.RuleName
}

// SetName sets the name of the ComputeResourceRule.
func (r *ComputeResourceRule) SetName(name string) {
	r.RuleName = name
}

// PrivilegeValidationRule defines the privilege validation rule
type PrivilegeValidationRule struct {
	validationrule.ManuallyNamed `json:",inline" yaml:",omitempty"`

	// RuleName is the name of the entity privilege validation rule
	RuleName string `json:"name" yaml:"name"`

	// Username is the username to validate against
	Username string `json:"username" yaml:"username"`

	// ClusterName is required when the vCenter Entity resides beneath a Cluster in the vCenter object hierarchy
	ClusterName string `json:"clusterName,omitempty" yaml:"clusterName,omitempty"`

	// EntityType is the type of the entity to validate
	// +kubebuilder:validation:Enum=cluster;datacenter;datastore;folder;host;network;resourcepool;vapp;vcenterroot;vds;vm
	EntityType string `json:"entityType" yaml:"entityType"`

	// EntityName is the name of the entity to validate
	EntityName string `json:"entityName" yaml:"entityName"`

	// Privileges is the list of privileges to validate that the user has
	Privileges []string `json:"privileges" yaml:"privileges"`

	// TODO: consider propagation somehow
}

var _ validationrule.Interface = (*PrivilegeValidationRule)(nil)

// Name returns the name of the EntityPrivilegeValidationRule.
func (r PrivilegeValidationRule) Name() string {
	return r.RuleName
}

// SetName sets the name of the EntityPrivilegeValidationRule.
func (r *PrivilegeValidationRule) SetName(name string) {
	r.RuleName = name
}

// TagValidationRule defines the tag validation rule
type TagValidationRule struct {
	validationrule.ManuallyNamed `json:",inline" yaml:",omitempty"`

	// RuleName is the name of the tag validation rule
	RuleName string `json:"name" yaml:"name"`

	// ClusterName is required when the vCenter Entity resides beneath a Cluster in the vCenter object hierarchy
	ClusterName string `json:"clusterName,omitempty" yaml:"clusterName"`

	// EntityType is the type of the entity to validate
	// +kubebuilder:validation:Enum=cluster;datacenter;folder;host;resourcepool;vm
	EntityType string `json:"entityType" yaml:"entityType"`

	// EntityName is the name of the entity to validate
	EntityName string `json:"entityName" yaml:"entityName"`

	// Tag is the tag to validate on the entity
	Tag string `json:"tag" yaml:"tag"`
}

var _ validationrule.Interface = (*TagValidationRule)(nil)

// Name returns the name of the TagValidationRule.
func (r TagValidationRule) Name() string {
	return r.RuleName
}

// SetName sets the name of the TagValidationRule.
func (r *TagValidationRule) SetName(name string) {
	r.RuleName = name
}

// NodepoolResourceRequirement defines the resource requirements for a nodepool
type NodepoolResourceRequirement struct {
	// Name is the name of the nodepool
	Name string `json:"name" yaml:"name"`

	// NumberOfNodes is the number of nodes in the nodepool
	NumberOfNodes int `json:"numberOfNodes" yaml:"numberOfNodes"`

	// CPU is the CPU requirement for the nodepool
	CPU string `json:"cpu" yaml:"cpu"`

	// Memory is the memory requirement for the nodepool
	Memory string `json:"memory" yaml:"memory"`

	// DiskSpace is the disk space requirement for the nodepool
	DiskSpace string `json:"diskSpace" yaml:"diskSpace"`
}

// VsphereValidatorStatus defines the observed state of VsphereValidator
type VsphereValidatorStatus struct{}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// VsphereValidator is the Schema for the vspherevalidators API
type VsphereValidator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VsphereValidatorSpec   `json:"spec,omitempty"`
	Status VsphereValidatorStatus `json:"status,omitempty"`
}

// GetKind returns the vSphere validator's kind.
func (v VsphereValidator) GetKind() string {
	return reflect.TypeOf(v).Name()
}

// PluginCode returns the vSphere validator's plugin code.
func (v VsphereValidator) PluginCode() string {
	return v.Spec.PluginCode()
}

// ResultCount returns the number of validation results expected for a VsphereValidator.
func (v VsphereValidator) ResultCount() int {
	return v.Spec.ResultCount()
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
