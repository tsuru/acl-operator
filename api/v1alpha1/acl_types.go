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
	TsuruApp      string                `json:"tsuruApp,omitempty"`
	RpaasInstance *ACLSpecRpaasInstance `json:"rpaasInstance,omitempty"`
}

type ACLSpecRpaasInstance struct {
	ServiceName string `json:"serviceName"`
	Instance    string `json:"instance"`
}

type ACLSpecDestination struct {
	TsuruApp      string                `json:"tsuruApp,omitempty"`
	TsuruAppPool  string                `json:"tsuruAppPool,omitempty"`
	RpaasInstance *ACLSpecRpaasInstance `json:"rpaasInstance,omitempty"`
	ExternalDNS   *ACLSpecExternalDNS   `json:"externalDNS,omitempty"`
	ExternalIP    *ACLSpecExternalIP    `json:"externalIP,omitempty"`
}

type ACLSpecExternalDNS struct {
	Name  string            `json:"name"`
	Ports ACLSpecProtoPorts `json:"ports,omitempty"`
}

type ACLSpecExternalIP struct {
	IP    string            `json:"ip"`
	Ports ACLSpecProtoPorts `json:"ports,omitempty"`
}

type ACLSpecProtoPorts []ProtoPort

type ProtoPort struct {
	Protocol string `json:"protocol"`
	Number   uint16 `json:"number"`
}

// ACLStatus defines the observed state of ACL
type ACLStatus struct {
	NetworkPolicy string   `json:"networkPolicy,omitempty"`
	Ready         bool     `json:"ready"`
	Reason        string   `json:"reason,omitempty"`
	WarningErrors []string `json:"warningErrors,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Ready",type=boolean,JSONPath=`.status.ready`

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
