//go:build !ignore_autogenerated

/*
Copyright 2024.

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
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ArgoAddon) DeepCopyInto(out *ArgoAddon) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ArgoAddon.
func (in *ArgoAddon) DeepCopy() *ArgoAddon {
	if in == nil {
		return nil
	}
	out := new(ArgoAddon)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ArgoAddon) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ArgoAddonList) DeepCopyInto(out *ArgoAddonList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ArgoAddon, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ArgoAddonList.
func (in *ArgoAddonList) DeepCopy() *ArgoAddonList {
	if in == nil {
		return nil
	}
	out := new(ArgoAddonList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ArgoAddonList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ArgoAddonSpec) DeepCopyInto(out *ArgoAddonSpec) {
	*out = *in
	out.Proxy = in.Proxy
	out.Argo = in.Argo
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ArgoAddonSpec.
func (in *ArgoAddonSpec) DeepCopy() *ArgoAddonSpec {
	if in == nil {
		return nil
	}
	out := new(ArgoAddonSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ArgoAddonStatus) DeepCopyInto(out *ArgoAddonStatus) {
	*out = *in
	out.Config = in.Config
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ArgoAddonStatus.
func (in *ArgoAddonStatus) DeepCopy() *ArgoAddonStatus {
	if in == nil {
		return nil
	}
	out := new(ArgoAddonStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ArgoTranslator) DeepCopyInto(out *ArgoTranslator) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ArgoTranslator.
func (in *ArgoTranslator) DeepCopy() *ArgoTranslator {
	if in == nil {
		return nil
	}
	out := new(ArgoTranslator)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ArgoTranslator) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ArgoTranslatorList) DeepCopyInto(out *ArgoTranslatorList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ArgoTranslator, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ArgoTranslatorList.
func (in *ArgoTranslatorList) DeepCopy() *ArgoTranslatorList {
	if in == nil {
		return nil
	}
	out := new(ArgoTranslatorList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ArgoTranslatorList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ArgoTranslatorSpec) DeepCopyInto(out *ArgoTranslatorSpec) {
	*out = *in
	if in.Selector != nil {
		in, out := &in.Selector, &out.Selector
		*out = new(v1.LabelSelector)
		(*in).DeepCopyInto(*out)
	}
	if in.ProjectRoles != nil {
		in, out := &in.ProjectRoles, &out.ProjectRoles
		*out = make([]ArgocdProjectRolesTranslator, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.ProjectSettings.DeepCopyInto(&out.ProjectSettings)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ArgoTranslatorSpec.
func (in *ArgoTranslatorSpec) DeepCopy() *ArgoTranslatorSpec {
	if in == nil {
		return nil
	}
	out := new(ArgoTranslatorSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ArgoTranslatorStatus) DeepCopyInto(out *ArgoTranslatorStatus) {
	*out = *in
	if in.Tenants != nil {
		in, out := &in.Tenants, &out.Tenants
		*out = make([]TenantStatus, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ArgoTranslatorStatus.
func (in *ArgoTranslatorStatus) DeepCopy() *ArgoTranslatorStatus {
	if in == nil {
		return nil
	}
	out := new(ArgoTranslatorStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ArgocdPolicyDefinition) DeepCopyInto(out *ArgocdPolicyDefinition) {
	*out = *in
	if in.Action != nil {
		in, out := &in.Action, &out.Action
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ArgocdPolicyDefinition.
func (in *ArgocdPolicyDefinition) DeepCopy() *ArgocdPolicyDefinition {
	if in == nil {
		return nil
	}
	out := new(ArgocdPolicyDefinition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ArgocdProjectPropertieMeta) DeepCopyInto(out *ArgocdProjectPropertieMeta) {
	*out = *in
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Annotations != nil {
		in, out := &in.Annotations, &out.Annotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Finalizers != nil {
		in, out := &in.Finalizers, &out.Finalizers
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ArgocdProjectPropertieMeta.
func (in *ArgocdProjectPropertieMeta) DeepCopy() *ArgocdProjectPropertieMeta {
	if in == nil {
		return nil
	}
	out := new(ArgocdProjectPropertieMeta)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ArgocdProjectProperties) DeepCopyInto(out *ArgocdProjectProperties) {
	*out = *in
	in.Structured.DeepCopyInto(&out.Structured)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ArgocdProjectProperties.
func (in *ArgocdProjectProperties) DeepCopy() *ArgocdProjectProperties {
	if in == nil {
		return nil
	}
	out := new(ArgocdProjectProperties)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ArgocdProjectRolesTranslator) DeepCopyInto(out *ArgocdProjectRolesTranslator) {
	*out = *in
	if in.ClusterRoles != nil {
		in, out := &in.ClusterRoles, &out.ClusterRoles
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Policies != nil {
		in, out := &in.Policies, &out.Policies
		*out = make([]ArgocdPolicyDefinition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ArgocdProjectRolesTranslator.
func (in *ArgocdProjectRolesTranslator) DeepCopy() *ArgocdProjectRolesTranslator {
	if in == nil {
		return nil
	}
	out := new(ArgocdProjectRolesTranslator)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ArgocdProjectStructuredProperties) DeepCopyInto(out *ArgocdProjectStructuredProperties) {
	*out = *in
	in.ProjectMeta.DeepCopyInto(&out.ProjectMeta)
	in.ProjectSpec.DeepCopyInto(&out.ProjectSpec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ArgocdProjectStructuredProperties.
func (in *ArgocdProjectStructuredProperties) DeepCopy() *ArgocdProjectStructuredProperties {
	if in == nil {
		return nil
	}
	out := new(ArgocdProjectStructuredProperties)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ControllerArgoCDConfig) DeepCopyInto(out *ControllerArgoCDConfig) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ControllerArgoCDConfig.
func (in *ControllerArgoCDConfig) DeepCopy() *ControllerArgoCDConfig {
	if in == nil {
		return nil
	}
	out := new(ControllerArgoCDConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ControllerCapsuleProxyConfig) DeepCopyInto(out *ControllerCapsuleProxyConfig) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ControllerCapsuleProxyConfig.
func (in *ControllerCapsuleProxyConfig) DeepCopy() *ControllerCapsuleProxyConfig {
	if in == nil {
		return nil
	}
	out := new(ControllerCapsuleProxyConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TenantStatus) DeepCopyInto(out *TenantStatus) {
	*out = *in
	in.Condition.DeepCopyInto(&out.Condition)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TenantStatus.
func (in *TenantStatus) DeepCopy() *TenantStatus {
	if in == nil {
		return nil
	}
	out := new(TenantStatus)
	in.DeepCopyInto(out)
	return out
}
