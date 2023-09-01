//go:build !ignore_autogenerated
// +build !ignore_autogenerated

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
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CloudAccountValidationRule) DeepCopyInto(out *CloudAccountValidationRule) {
	*out = *in
	if in.Expressions != nil {
		in, out := &in.Expressions, &out.Expressions
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CloudAccountValidationRule.
func (in *CloudAccountValidationRule) DeepCopy() *CloudAccountValidationRule {
	if in == nil {
		return nil
	}
	out := new(CloudAccountValidationRule)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DiskSpaceValidationRule) DeepCopyInto(out *DiskSpaceValidationRule) {
	*out = *in
	if in.Expressions != nil {
		in, out := &in.Expressions, &out.Expressions
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DiskSpaceValidationRule.
func (in *DiskSpaceValidationRule) DeepCopy() *DiskSpaceValidationRule {
	if in == nil {
		return nil
	}
	out := new(DiskSpaceValidationRule)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RegionZoneValidationRule) DeepCopyInto(out *RegionZoneValidationRule) {
	*out = *in
	if in.Clusters != nil {
		in, out := &in.Clusters, &out.Clusters
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RegionZoneValidationRule.
func (in *RegionZoneValidationRule) DeepCopy() *RegionZoneValidationRule {
	if in == nil {
		return nil
	}
	out := new(RegionZoneValidationRule)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RolePrivilegeValidationRule) DeepCopyInto(out *RolePrivilegeValidationRule) {
	*out = *in
	if in.Expressions != nil {
		in, out := &in.Expressions, &out.Expressions
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RolePrivilegeValidationRule.
func (in *RolePrivilegeValidationRule) DeepCopy() *RolePrivilegeValidationRule {
	if in == nil {
		return nil
	}
	out := new(RolePrivilegeValidationRule)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VsphereAuth) DeepCopyInto(out *VsphereAuth) {
	*out = *in
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
	out.Auth = in.Auth
	if in.RolePrivilegeValidationRules != nil {
		in, out := &in.RolePrivilegeValidationRules, &out.RolePrivilegeValidationRules
		*out = make([]RolePrivilegeValidationRule, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.RegionZoneValidationRule.DeepCopyInto(&out.RegionZoneValidationRule)
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
