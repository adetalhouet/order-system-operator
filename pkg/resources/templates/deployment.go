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

// GetDeploymentName returns the name of the deployment
func GetDeploymentName(orderSystem *appsv1alpha1.OrderSystem, podName string) string {
	return podName + "-" + orderSystem.Name
}

// DeploymentSpec is the deployment manifest template
func DeploymentSpec(orderSystem *appsv1alpha1.OrderSystem, deploymentName string, podName string, service Service) *appsv1.Deployment {
	isIstioEnabled := orderSystem.Spec.InjectIstioSidecarEnabled
	ls := GetOrderSystemLabels(orderSystem.Name)
	cm := corev1.LocalObjectReference{
		Name: ConfigMapApplicationConfigurationName,
	}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
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
						Image:           "adetalhouet/order-system-" + podName + ":" + orderSystem.Spec.Version,
						ImagePullPolicy: corev1.PullAlways,
						Name:            podName,
						Ports: []corev1.ContainerPort{{
							ContainerPort: service.Port,
							Name:          strings.Split(deploymentName, "-")[0] + "-http",
						}},
						Env: []corev1.EnvVar{
							genEnvFromSecret("DB_USERNAME", orderSystem.Spec.DbInfo.Secret, "username"),
							genEnvFromSecret("DB_PASSWORD", orderSystem.Spec.DbInfo.Secret, "password"),
							genEnvFromSecret("NATS_USERNAME", orderSystem.Spec.DbInfo.Secret, "username"),
							genEnvFromSecret("NATS_PASSWORD", orderSystem.Spec.DbInfo.Secret, "password"),
						},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "application-conf",
								MountPath: "/app/resources/application.conf",
								SubPath:   "application.conf",
							},
						},
					}},
					Volumes: []corev1.Volume{
						{
							Name: "application-conf",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: cm,
								},
							},
						},
					},
				},
			},
		},
	}
	return dep
}

func getAnnotations(isIstioEnabled bool) map[string]string {
	return map[string]string{isIstioEnabledAnnotation: strconv.FormatBool(isIstioEnabled)}
}

func genEnvFromSecret(envName string, secretName string, secretKey string) corev1.EnvVar {
	env := corev1.EnvVar{
		Name: envName,
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: secretName,
				},
				Key: secretKey,
			},
		},
	}
	return env
}
