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
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appsv1alpha1 "github.com/adetalhouet/order-system-operator/api/v1alpha1"
)

func GetOrderSystemLabels(name string) map[string]string {
	return map[string]string{"app": "order-system", "ordersystem_cr": name}
}

func DeploymentList(orderSystem *appsv1alpha1.OrderSystem) []appsv1.Deployment {

	deps := make([]appsv1.Deployment, 0)

	deps = append(deps, deploymentSpec(orderSystem, "adetalhouet/cart-service", "cart-service", 9090))
	deps = append(deps, deploymentSpec(orderSystem, "adetalhouet/client-service", "client-service", 9091))
	deps = append(deps, deploymentSpec(orderSystem, "adetalhouet/order-service", "order-service", 9092))
	deps = append(deps, deploymentSpec(orderSystem, "adetalhouet/product-service", "product-service", 9093))
	deps = append(deps, deploymentSpec(orderSystem, "adetalhouet/order-system-api-gw", "order-system-api-gw", 9090))

	return deps
}

func deploymentSpec(orderSystem *appsv1alpha1.OrderSystem, containerName string, deployementName string, port int32) appsv1.Deployment {
	isIstioEnabled := orderSystem.Spec.InjectIstioSidecarEnabled
	ls := GetOrderSystemLabels(orderSystem.Name)

	dep := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deployementName + "-" + orderSystem.Name,
			Namespace: orderSystem.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      ls,
					Annotations: getOrderSystemAnnotations(isIstioEnabled),
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image:           containerName + ":" + orderSystem.Spec.Version,
						ImagePullPolicy: corev1.PullAlways,
						Name:            deployementName,
						Ports: []corev1.ContainerPort{{
							ContainerPort: port,
							Name:          strings.Split(deployementName, "-")[0] + "-http",
						}},
					}},
				},
			},
		},
	}
	return dep
}

func getOrderSystemAnnotations(isIstioEnabled string) map[string]string {
	return map[string]string{"sidecar.istio.io/inject": isIstioEnabled}
}
