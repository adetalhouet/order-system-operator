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

package templates

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appsv1alpha1 "github.com/adetalhouet/order-system-operator/api/v1alpha1"
)

type Service struct {
	Name string
	Port int32
	Type corev1.ServiceType
}

// ServiceSpec is the service manifest template
func ServiceSpec(orderSystem *appsv1alpha1.OrderSystem, deploymentName string, service Service) *corev1.Service {
	selector := map[string]string{
		"app": deploymentName,
	}
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName + "-svc",
			Namespace: orderSystem.Namespace,
			Labels:    GetOrderSystemLabels(orderSystem.Name),
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:     deploymentName,
					Protocol: "TCP",
					Port:     service.Port,
				},
			},
			// ExternalIPs: []string{
			// orderSystem.Spec.ExternalIP,
			// },
			Type:     service.Type,
			Selector: selector,
		},
	}
}
