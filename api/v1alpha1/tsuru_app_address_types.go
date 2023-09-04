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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TsuruAppAddressSpec defines the desired state of TsuruAppAddress
type TsuruAppAddressSpec struct {
	Name          string   `json:"name,omitempty"`
	AdditionalIPs []string `json:"additionalIPs,omitempty"`
}

// ResourceAddressStatus defines the observed state of TsuruAppAddress and RpaasInstanceAddress
type ResourceAddressStatus struct {
	Ready     bool     `json:"ready"`
	Reason    string   `json:"reason,omitempty"`
	UpdatedAt string   `json:"updatedAt,omitempty"`
	IPs       []string `json:"ips,omitempty"`
	Pool      string   `json:"pool,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:printcolumn:name="Ready",type=boolean,JSONPath=`.status.ready`
//+kubebuilder:printcolumn:name="Addresses",type=string,JSONPath=`.status.ips[*]`

// TsuruAppAddress is the Schema for the tsuruappaddresses API
type TsuruAppAddress struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TsuruAppAddressSpec   `json:"spec,omitempty"`
	Status ResourceAddressStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TsuruAppAddressList contains a list of TsuruAppAddress
type TsuruAppAddressList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TsuruAppAddress `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TsuruAppAddress{}, &TsuruAppAddressList{})
}
