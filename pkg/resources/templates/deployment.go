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
	"strconv"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appsv1alpha1 "github.com/adetalhouet/order-system-operator/api/v1alpha1"
)

const isIstioEnabledAnnotation = "sidecar.istio.io/inject"

// IsAnnotationsValidOrSet validates current deployment annotations are as define as expected,
// else set them properly. Return `true` is annotations were valid, `false` otherwise
func IsAnnotationsValidOrSet(currentDep *appsv1.Deployment, value bool) bool {
	if currentDep.GetAnnotations()[isIstioEnabledAnnotation] != strconv.FormatBool(value) {
		currentDep.SetAnnotations(getAnnotations(value))
		return false
	}
	return true
}

// GetOrderSystemLabels returns the labels for that CRD deployments
func GetOrderSystemLabels(name string) map[string]string {
	return map[string]string{"app": "order-system", "ordersystem_cr": name}
}

// IsOrderSystemLabels checks whether we deal with order system object
func IsOrderSystemLabels(labels map[string]string) bool {
	if labels != nil && labels["app"] == "order-system" {
		return true
	}
	return false
}

// GetDeploymentName returns the name of the deployment
func GetDeploymentName(orderSystem *appsv1alpha1.OrderSystem, deploymentName string) string {
	return deploymentName + "-" + orderSystem.Name
}

// DeploymentSpec is the deployment manifest template
func DeploymentSpec(orderSystem *appsv1alpha1.OrderSystem, deploymentName string, port int32) *appsv1.Deployment {
	isIstioEnabled := orderSystem.Spec.InjectIstioSidecarEnabled
	ls := GetOrderSystemLabels(orderSystem.Name)

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetDeploymentName(orderSystem, deploymentName),
			Namespace: orderSystem.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      ls,
					Annotations: getAnnotations(isIstioEnabled),
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image:           "adetalhouet/" + deploymentName + ":" + orderSystem.Spec.Version,
						ImagePullPolicy: corev1.PullAlways,
						Name:            deploymentName,
						Ports: []corev1.ContainerPort{{
							ContainerPort: port,
							Name:          strings.Split(deploymentName, "-")[0] + "-http",
						}},
					}},
				},
			},
		},
	}
	return dep
}

func getAnnotations(isIstioEnabled bool) map[string]string {
	return map[string]string{isIstioEnabledAnnotation: strconv.FormatBool(isIstioEnabled)}
}
