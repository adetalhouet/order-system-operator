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
	"fmt"
	"strings"

	appsv1alpha1 "github.com/adetalhouet/order-system-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConfigMapApplicationConfigurationName is the name of the application configuration configmap
const ConfigMapApplicationConfigurationName = "order-system-apps-config"

// ConfigMapSpec returns the configmap manifest
func ConfigMapSpec(orderSystem *appsv1alpha1.OrderSystem, apps map[string]Service, dbServiceName string, natServiceName string) *corev1.ConfigMap {
	configMapData := make(map[string]string, 0)
	configMapData["application.conf"] = buildConfig(apps, dbServiceName, natServiceName)
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ConfigMapApplicationConfigurationName,
			Namespace: orderSystem.Namespace,
		},
		Data: configMapData,
	}
}

// BuildConfig builds the config for the apps
// TODO move creds to secret
func buildConfig(apps map[string]Service, dbServiceName string, natServiceName string) Config {
	var configTemplate = `
	nats {
		host = "nats://natServiceName:4222"
		connectionTimeout = 10
		pingInterval = 20
		maxPingsOut = 5
		maxReconnects = 10
		reconnectWait = 10
		connectionName = "Test Order System NATS bus"
	  
		username = "order-system"
		password = "Password123"
	  }
	  postgres {
		driverName = "org.h2.Driver"
		url = "jdbc:postgresql://dbServiceName:5432/order-system"
		username = "order-system"
		password = "Password123"
	  }
	  order {
		url = "order-url"
		port = order-port
	  }
	  product {
		url = "product-url"
		port = product-port
	  }
	  client {
		url = "client-url"
		port = client-port
	  }
	  cart {
		url = "cart-url"
		port = cart-port
	  }
	  api-gw {
		port = api-gw-port
	  }
	  `

	config := strings.Replace(configTemplate, "dbServiceName", dbServiceName, -1)
	config = strings.Replace(config, "natServiceName", natServiceName, -1)

	for podName, service := range apps {
		app := strings.Split(podName, "-service")[0]
		r := strings.NewReplacer(
			app+"-url", service.Name,
			app+"-port", fmt.Sprint(service.Port))

		config = r.Replace(configTemplate)
	}

	return config
}
