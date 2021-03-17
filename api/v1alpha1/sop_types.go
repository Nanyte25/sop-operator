/*
Copyright 2021.

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

// SOPSpec defines the desired state of SOP
type SOPSpec struct {
	Identifier string `json:"identifier"`
}

// SOPStatus defines the observed state of SOP
type SOPStatus struct {
	Phase string `json:"phase"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// SOP is the Schema for the sops API
type SOP struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SOPSpec   `json:"spec,omitempty"`
	Status SOPStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SOPList contains a list of SOP
type SOPList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SOP `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SOP{}, &SOPList{})
}
