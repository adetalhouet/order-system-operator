/*
Copyright 2020 Alexis de Talhouët.

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
	"github.com/operator-framework/operator-sdk/pkg/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Info struct {
	// Secret containing credentials, in form of username and password
	Secret string `json:"credentialSecretName"`
	// Service name
	Service string `json:"serviceName"`
}

// OrderSystemSpec defines the desired state of OrderSystem
type OrderSystemSpec struct {
	// Version of the Order System
	Version string `json:"version"`
	// Whether or not to inject Istio
	InjectIstioSidecarEnabled bool `json:"injectIstioSidecarEnabled"`
	// Autoscale
	AutoscaleEnabled bool `json:"autoscaleEnabled"`
	// Database service
	DbInfo Info `json:"postgres"`
	// Nats service
	NatsInfo Info `json:"nats"`
}

// OrderSystemStatus defines the observed state of OrderSystem
type OrderSystemStatus struct {
	Conditions status.Conditions `json:"conditions"`
}

func (orderSystem *OrderSystem) GetReconcileStatus() status.Conditions {
	return orderSystem.Status.Conditions
}

func (orderSystem *OrderSystem) SetReconcileStatus(reconcileStatus status.Conditions) {
	orderSystem.Status.Conditions = reconcileStatus
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// OrderSystem is the Schema for the ordersystems API
type OrderSystem struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OrderSystemSpec   `json:"spec,omitempty"`
	Status OrderSystemStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// OrderSystemList contains a list of OrderSystem
type OrderSystemList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OrderSystem `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OrderSystem{}, &OrderSystemList{})
}
