//go:build !ignore_autogenerated

/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"github.com/validator-labs/validator-plugin-vsphere/api/vcenter"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ComputeResourceRule) DeepCopyInto(out *ComputeResourceRule) {
	*out = *in
	out.ManuallyNamed = in.ManuallyNamed
	if in.NodepoolResourceRequirements != nil {
		in, out := &in.NodepoolResourceRequirements, &out.NodepoolResourceRequirements
		*out = make([]NodepoolResourceRequirement, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ComputeResourceRule.
func (in *ComputeResourceRule) DeepCopy() *ComputeResourceRule {
	if in == nil {
		return nil
	}
	out := new(ComputeResourceRule)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NTPValidationRule) DeepCopyInto(out *NTPValidationRule) {
	*out = *in
	out.ManuallyNamed = in.ManuallyNamed
	if in.Hosts != nil {
		in, out := &in.Hosts, &out.Hosts
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NTPValidationRule.
func (in *NTPValidationRule) DeepCopy() *NTPValidationRule {
	if in == nil {
		return nil
	}
	out := new(NTPValidationRule)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodepoolResourceRequirement) DeepCopyInto(out *NodepoolResourceRequirement) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodepoolResourceRequirement.
func (in *NodepoolResourceRequirement) DeepCopy() *NodepoolResourceRequirement {
	if in == nil {
		return nil
	}
	out := new(NodepoolResourceRequirement)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PrivilegeValidationRule) DeepCopyInto(out *PrivilegeValidationRule) {
	*out = *in
	out.ManuallyNamed = in.ManuallyNamed
	if in.Privileges != nil {
		in, out := &in.Privileges, &out.Privileges
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	in.Propagation.DeepCopyInto(&out.Propagation)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PrivilegeValidationRule.
func (in *PrivilegeValidationRule) DeepCopy() *PrivilegeValidationRule {
	if in == nil {
		return nil
	}
	out := new(PrivilegeValidationRule)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Propagation) DeepCopyInto(out *Propagation) {
	*out = *in
	if in.GroupPrincipals != nil {
		in, out := &in.GroupPrincipals, &out.GroupPrincipals
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Propagation.
func (in *Propagation) DeepCopy() *Propagation {
	if in == nil {
		return nil
	}
	out := new(Propagation)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TagValidationRule) DeepCopyInto(out *TagValidationRule) {
	*out = *in
	out.ManuallyNamed = in.ManuallyNamed
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TagValidationRule.
func (in *TagValidationRule) DeepCopy() *TagValidationRule {
	if in == nil {
		return nil
	}
	out := new(TagValidationRule)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VsphereAuth) DeepCopyInto(out *VsphereAuth) {
	*out = *in
	if in.Account != nil {
		in, out := &in.Account, &out.Account
		*out = new(vcenter.Account)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VsphereAuth.
func (in *VsphereAuth) DeepCopy() *VsphereAuth {
	if in == nil {
		return nil
	}
	out := new(VsphereAuth)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VsphereValidator) DeepCopyInto(out *VsphereValidator) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VsphereValidator.
func (in *VsphereValidator) DeepCopy() *VsphereValidator {
	if in == nil {
		return nil
	}
	out := new(VsphereValidator)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VsphereValidator) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VsphereValidatorList) DeepCopyInto(out *VsphereValidatorList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]VsphereValidator, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VsphereValidatorList.
func (in *VsphereValidatorList) DeepCopy() *VsphereValidatorList {
	if in == nil {
		return nil
	}
	out := new(VsphereValidatorList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VsphereValidatorList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VsphereValidatorSpec) DeepCopyInto(out *VsphereValidatorSpec) {
	*out = *in
	in.Auth.DeepCopyInto(&out.Auth)
	if in.PrivilegeValidationRules != nil {
		in, out := &in.PrivilegeValidationRules, &out.PrivilegeValidationRules
		*out = make([]PrivilegeValidationRule, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.TagValidationRules != nil {
		in, out := &in.TagValidationRules, &out.TagValidationRules
		*out = make([]TagValidationRule, len(*in))
		copy(*out, *in)
	}
	if in.ComputeResourceRules != nil {
		in, out := &in.ComputeResourceRules, &out.ComputeResourceRules
		*out = make([]ComputeResourceRule, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.NTPValidationRules != nil {
		in, out := &in.NTPValidationRules, &out.NTPValidationRules
		*out = make([]NTPValidationRule, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VsphereValidatorSpec.
func (in *VsphereValidatorSpec) DeepCopy() *VsphereValidatorSpec {
	if in == nil {
		return nil
	}
	out := new(VsphereValidatorSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VsphereValidatorStatus) DeepCopyInto(out *VsphereValidatorStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VsphereValidatorStatus.
func (in *VsphereValidatorStatus) DeepCopy() *VsphereValidatorStatus {
	if in == nil {
		return nil
	}
	out := new(VsphereValidatorStatus)
	in.DeepCopyInto(out)
	return out
}
