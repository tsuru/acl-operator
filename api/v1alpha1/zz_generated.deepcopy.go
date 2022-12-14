//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright 2022.

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
	"k8s.io/api/networking/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ACL) DeepCopyInto(out *ACL) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ACL.
func (in *ACL) DeepCopy() *ACL {
	if in == nil {
		return nil
	}
	out := new(ACL)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ACL) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ACLDNSEntry) DeepCopyInto(out *ACLDNSEntry) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ACLDNSEntry.
func (in *ACLDNSEntry) DeepCopy() *ACLDNSEntry {
	if in == nil {
		return nil
	}
	out := new(ACLDNSEntry)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ACLDNSEntry) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ACLDNSEntryList) DeepCopyInto(out *ACLDNSEntryList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ACLDNSEntry, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ACLDNSEntryList.
func (in *ACLDNSEntryList) DeepCopy() *ACLDNSEntryList {
	if in == nil {
		return nil
	}
	out := new(ACLDNSEntryList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ACLDNSEntryList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ACLDNSEntrySpec) DeepCopyInto(out *ACLDNSEntrySpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ACLDNSEntrySpec.
func (in *ACLDNSEntrySpec) DeepCopy() *ACLDNSEntrySpec {
	if in == nil {
		return nil
	}
	out := new(ACLDNSEntrySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ACLDNSEntryStatus) DeepCopyInto(out *ACLDNSEntryStatus) {
	*out = *in
	if in.IPs != nil {
		in, out := &in.IPs, &out.IPs
		*out = make([]ACLDNSEntryStatusIP, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ACLDNSEntryStatus.
func (in *ACLDNSEntryStatus) DeepCopy() *ACLDNSEntryStatus {
	if in == nil {
		return nil
	}
	out := new(ACLDNSEntryStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ACLDNSEntryStatusIP) DeepCopyInto(out *ACLDNSEntryStatusIP) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ACLDNSEntryStatusIP.
func (in *ACLDNSEntryStatusIP) DeepCopy() *ACLDNSEntryStatusIP {
	if in == nil {
		return nil
	}
	out := new(ACLDNSEntryStatusIP)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ACLList) DeepCopyInto(out *ACLList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ACL, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ACLList.
func (in *ACLList) DeepCopy() *ACLList {
	if in == nil {
		return nil
	}
	out := new(ACLList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ACLList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ACLSpec) DeepCopyInto(out *ACLSpec) {
	*out = *in
	in.Source.DeepCopyInto(&out.Source)
	if in.Destinations != nil {
		in, out := &in.Destinations, &out.Destinations
		*out = make([]ACLSpecDestination, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ACLSpec.
func (in *ACLSpec) DeepCopy() *ACLSpec {
	if in == nil {
		return nil
	}
	out := new(ACLSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ACLSpecDestination) DeepCopyInto(out *ACLSpecDestination) {
	*out = *in
	if in.RpaasInstance != nil {
		in, out := &in.RpaasInstance, &out.RpaasInstance
		*out = new(ACLSpecRpaasInstance)
		**out = **in
	}
	if in.ExternalDNS != nil {
		in, out := &in.ExternalDNS, &out.ExternalDNS
		*out = new(ACLSpecExternalDNS)
		(*in).DeepCopyInto(*out)
	}
	if in.ExternalIP != nil {
		in, out := &in.ExternalIP, &out.ExternalIP
		*out = new(ACLSpecExternalIP)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ACLSpecDestination.
func (in *ACLSpecDestination) DeepCopy() *ACLSpecDestination {
	if in == nil {
		return nil
	}
	out := new(ACLSpecDestination)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ACLSpecExternalDNS) DeepCopyInto(out *ACLSpecExternalDNS) {
	*out = *in
	if in.Ports != nil {
		in, out := &in.Ports, &out.Ports
		*out = make(ACLSpecProtoPorts, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ACLSpecExternalDNS.
func (in *ACLSpecExternalDNS) DeepCopy() *ACLSpecExternalDNS {
	if in == nil {
		return nil
	}
	out := new(ACLSpecExternalDNS)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ACLSpecExternalIP) DeepCopyInto(out *ACLSpecExternalIP) {
	*out = *in
	if in.Ports != nil {
		in, out := &in.Ports, &out.Ports
		*out = make(ACLSpecProtoPorts, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ACLSpecExternalIP.
func (in *ACLSpecExternalIP) DeepCopy() *ACLSpecExternalIP {
	if in == nil {
		return nil
	}
	out := new(ACLSpecExternalIP)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in ACLSpecProtoPorts) DeepCopyInto(out *ACLSpecProtoPorts) {
	{
		in := &in
		*out = make(ACLSpecProtoPorts, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ACLSpecProtoPorts.
func (in ACLSpecProtoPorts) DeepCopy() ACLSpecProtoPorts {
	if in == nil {
		return nil
	}
	out := new(ACLSpecProtoPorts)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ACLSpecRpaasInstance) DeepCopyInto(out *ACLSpecRpaasInstance) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ACLSpecRpaasInstance.
func (in *ACLSpecRpaasInstance) DeepCopy() *ACLSpecRpaasInstance {
	if in == nil {
		return nil
	}
	out := new(ACLSpecRpaasInstance)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ACLSpecSource) DeepCopyInto(out *ACLSpecSource) {
	*out = *in
	if in.RpaasInstance != nil {
		in, out := &in.RpaasInstance, &out.RpaasInstance
		*out = new(ACLSpecRpaasInstance)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ACLSpecSource.
func (in *ACLSpecSource) DeepCopy() *ACLSpecSource {
	if in == nil {
		return nil
	}
	out := new(ACLSpecSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ACLStatus) DeepCopyInto(out *ACLStatus) {
	*out = *in
	if in.WarningErrors != nil {
		in, out := &in.WarningErrors, &out.WarningErrors
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Stale != nil {
		in, out := &in.Stale, &out.Stale
		*out = make([]ACLStatusStale, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.RuleErrors != nil {
		in, out := &in.RuleErrors, &out.RuleErrors
		*out = make([]ACLStatusRuleError, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ACLStatus.
func (in *ACLStatus) DeepCopy() *ACLStatus {
	if in == nil {
		return nil
	}
	out := new(ACLStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ACLStatusRuleError) DeepCopyInto(out *ACLStatusRuleError) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ACLStatusRuleError.
func (in *ACLStatusRuleError) DeepCopy() *ACLStatusRuleError {
	if in == nil {
		return nil
	}
	out := new(ACLStatusRuleError)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ACLStatusStale) DeepCopyInto(out *ACLStatusStale) {
	*out = *in
	if in.Rules != nil {
		in, out := &in.Rules, &out.Rules
		*out = make([]v1.NetworkPolicyEgressRule, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ACLStatusStale.
func (in *ACLStatusStale) DeepCopy() *ACLStatusStale {
	if in == nil {
		return nil
	}
	out := new(ACLStatusStale)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ProtoPort) DeepCopyInto(out *ProtoPort) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ProtoPort.
func (in *ProtoPort) DeepCopy() *ProtoPort {
	if in == nil {
		return nil
	}
	out := new(ProtoPort)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceAddressStatus) DeepCopyInto(out *ResourceAddressStatus) {
	*out = *in
	if in.IPs != nil {
		in, out := &in.IPs, &out.IPs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceAddressStatus.
func (in *ResourceAddressStatus) DeepCopy() *ResourceAddressStatus {
	if in == nil {
		return nil
	}
	out := new(ResourceAddressStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RpaasInstanceAddress) DeepCopyInto(out *RpaasInstanceAddress) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RpaasInstanceAddress.
func (in *RpaasInstanceAddress) DeepCopy() *RpaasInstanceAddress {
	if in == nil {
		return nil
	}
	out := new(RpaasInstanceAddress)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RpaasInstanceAddress) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RpaasInstanceAddressList) DeepCopyInto(out *RpaasInstanceAddressList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]RpaasInstanceAddress, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RpaasInstanceAddressList.
func (in *RpaasInstanceAddressList) DeepCopy() *RpaasInstanceAddressList {
	if in == nil {
		return nil
	}
	out := new(RpaasInstanceAddressList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RpaasInstanceAddressList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RpaasInstanceAddressSpec) DeepCopyInto(out *RpaasInstanceAddressSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RpaasInstanceAddressSpec.
func (in *RpaasInstanceAddressSpec) DeepCopy() *RpaasInstanceAddressSpec {
	if in == nil {
		return nil
	}
	out := new(RpaasInstanceAddressSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TsuruAppAddress) DeepCopyInto(out *TsuruAppAddress) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TsuruAppAddress.
func (in *TsuruAppAddress) DeepCopy() *TsuruAppAddress {
	if in == nil {
		return nil
	}
	out := new(TsuruAppAddress)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *TsuruAppAddress) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TsuruAppAddressList) DeepCopyInto(out *TsuruAppAddressList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]TsuruAppAddress, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TsuruAppAddressList.
func (in *TsuruAppAddressList) DeepCopy() *TsuruAppAddressList {
	if in == nil {
		return nil
	}
	out := new(TsuruAppAddressList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *TsuruAppAddressList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TsuruAppAddressSpec) DeepCopyInto(out *TsuruAppAddressSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TsuruAppAddressSpec.
func (in *TsuruAppAddressSpec) DeepCopy() *TsuruAppAddressSpec {
	if in == nil {
		return nil
	}
	out := new(TsuruAppAddressSpec)
	in.DeepCopyInto(out)
	return out
}
