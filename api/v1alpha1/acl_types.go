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

// ACLSpec defines the desired state of ACL
type ACLSpec struct {
	Source       ACLSpecSource        `json:"source"`
	Destinations []ACLSpecDestination `json:"destinations"`
}

type ACLSpecSource struct {
	TsuruApp      string                      `json:"tsuruApp"`
	RpaasInstance *ACLSpecSourceRpaasInstance `json:"rpaasInstance"`
}

type ACLSpecSourceRpaasInstance struct {
	ServiceName string `json:"serviceName"`
	Instance    string `json:"instance"`
}

type ACLSpecDestination struct {
	TsuruApp      string                      `json:"tsuruApp"`
	TsuruAppPool  string                      `json:"tsuruAppPool"`
	RpaasInstance *ACLSpecSourceRpaasInstance `json:"rpaasInstance"`
	ExternalDNS   *ACLSpecExternalDNS         `json:"externalDNS"`
	ExternalIP    *ACLSpecExternalIP          `json:"externalIP"`
}

type ACLSpecExternalDNS struct {
	Name  string            `json:"name"`
	Ports ACLSpecProtoPorts `json:"ports"`
}

type ACLSpecExternalIP struct {
	IP    string            `json:"ip"`
	Ports ACLSpecProtoPorts `json:"ports"`
}

type ACLSpecProtoPorts []ProtoPort

type ProtoPort struct {
	Protocol string `json:"protocol"`
	Port     uint16 `json:"port"`
}

// ACLStatus defines the observed state of ACL
type ACLStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ACL is the Schema for the acls API
type ACL struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ACLSpec   `json:"spec,omitempty"`
	Status ACLStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ACLList contains a list of ACL
type ACLList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ACL `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ACL{}, &ACLList{})
}
