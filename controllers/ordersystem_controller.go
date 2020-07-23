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

package controllers

import (
	"context"
	"reflect"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	appsv1alpha1 "github.com/adetalhouet/order-system-operator/api/v1alpha1"
	"github.com/adetalhouet/order-system-operator/pkg/resources/templates"
)

// OrderSystemReconciler reconciles a OrderSystem object
type OrderSystemReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=apps.adetalhouet.io,resources=ordersystems,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.adetalhouet.io,resources=ordersystems/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;

func (r *OrderSystemReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("ordersystem", req.NamespacedName)

	log.Info("Reconcile OrderSystem")

	// Fetch the OrderSystem instance
	orderSystem := &appsv1alpha1.OrderSystem{}
	if err := r.Get(ctx, req.NamespacedName, orderSystem); err != nil {
		log.Error(err, "Unable to get OrderSystem")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Check if the deployment already exists, if not create a new one
	found := &appsv1.Deployment{}
	if err := r.Get(ctx, types.NamespacedName{Name: orderSystem.Name, Namespace: orderSystem.Namespace}, found); err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		deps := templates.DeploymentList(orderSystem)

		log.Info("Deploying OrderSystem")

		for _, dep := range deps {
			log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			err = r.Create(ctx, &dep)
			if err != nil {
				log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
				return ctrl.Result{}, err
			}
		}
		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	// Ensure the deployment size is the same as the spec
	// size := orderSystem.Spec.Size
	// if *found.Spec.Replicas != size {
	// 	found.Spec.Replicas = &size
	// 	if err := r.Update(ctx, found); err != nil {
	// 		log.Error(err, "Failed to update Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
	// 		return ctrl.Result{}, err
	// 	}
	// 	// Spec updated - return and requeue
	// 	return ctrl.Result{Requeue: true}, nil
	// }

	// Update the OrderSystem status with the pod names
	// List the pods for this orderSystem's deployment
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(orderSystem.Namespace),
		client.MatchingLabels(templates.GetOrderSystemLabels(orderSystem.Name)),
	}
	if err := r.List(ctx, podList, listOpts...); err != nil {
		log.Error(err, "Failed to list pods", "OrderSystem.Namespace", orderSystem.Namespace, "OrderSystem.Name", orderSystem.Name)
		return ctrl.Result{}, err
	}
	podNames := getPodNames(podList.Items)

	// Update status.Nodes if needed
	if !reflect.DeepEqual(podNames, orderSystem.Status.Nodes) {
		orderSystem.Status.Nodes = podNames
		err := r.Status().Update(ctx, orderSystem)
		if err != nil {
			log.Error(err, "Failed to update Memcached status")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}

func (r *OrderSystemReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.OrderSystem{}).
		Owns(&appsv1.Deployment{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 2,
		}).
		Complete(r)
}
