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

// TsuruAppAdressSpec defines the desired state of TsuruAppAdress
type TsuruAppAdressSpec struct {
	Name string `json:"name,omitempty"`
}

// TsuruAppAdressStatus defines the observed state of TsuruAppAdress
type TsuruAppAdressStatus struct {
	Ready     bool     `json:"ready"`
	UpdatedAt string   `json:"updatedAt,omitempty"`
	RouterIPs []string `json:"routerIPs"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// TsuruAppAdress is the Schema for the tsuruappadresses API
type TsuruAppAdress struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TsuruAppAdressSpec   `json:"spec,omitempty"`
	Status TsuruAppAdressStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TsuruAppAdressList contains a list of TsuruAppAdress
type TsuruAppAdressList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TsuruAppAdress `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TsuruAppAdress{}, &TsuruAppAdressList{})
}