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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RpaasInstanceAdressSpec defines the desired state of RpaasInstanceAdress
type RpaasInstanceAdressSpec struct {
	ServiceName string `json:"serviceName,omitempty"`
	Instance    string `json:"instance,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:printcolumn:name="Ready",type=boolean,JSONPath=`.status.ready`
//+kubebuilder:printcolumn:name="Addresses",type=string,JSONPath=`.status.ips[*]`

// RpaasInstanceAdress is the Schema for the rpaasinstanceadresses API
type RpaasInstanceAdress struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RpaasInstanceAdressSpec `json:"spec,omitempty"`
	Status ResourceAdressStatus    `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RpaasInstanceAdressList contains a list of RpaasInstanceAdress
type RpaasInstanceAdressList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RpaasInstanceAdress `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RpaasInstanceAdress{}, &RpaasInstanceAdressList{})
}
