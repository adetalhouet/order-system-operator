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
	"errors"

	appsv1alpha1 "github.com/adetalhouet/order-system-operator/api/v1alpha1"
	"github.com/adetalhouet/order-system-operator/pkg/resources/templates"
	"github.com/redhat-cop/operator-utils/pkg/util"
	"github.com/redhat-cop/operator-utils/pkg/util/apis"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	types "k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const controllerName = "ordersystem_crd"
const orderSystemFinalizer = "io.adetalhouet.ordersystem.finalizer"

var log = logf.Log.WithName(controllerName)

// OrderSystemReconciler reconciles a OrderSystem object
type OrderSystemReconciler struct {
	util.ReconcilerBase
}

// Add creates a new orderSystem Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &OrderSystemReconciler{
		ReconcilerBase: util.NewReconcilerBase(mgr.GetClient(), mgr.GetScheme(), mgr.GetConfig(), mgr.GetEventRecorderFor(controllerName)),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New(controllerName, mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource OrderSystem
	err = c.Watch(&source.Kind{Type: &appsv1alpha1.OrderSystem{}}, &handler.EnqueueRequestForObject{}, util.ResourceGenerationOrFinalizerChangedPredicate{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource deployment OrderSystem
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appsv1alpha1.OrderSystem{},
	}, util.ResourceGenerationOrFinalizerChangedPredicate{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource service OrderSystem
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appsv1alpha1.OrderSystem{},
	}, util.ResourceGenerationOrFinalizerChangedPredicate{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource configmap OrderSystem
	err = c.Watch(&source.Kind{Type: &corev1.ConfigMap{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appsv1alpha1.OrderSystem{},
	}, util.ResourceGenerationOrFinalizerChangedPredicate{})
	if err != nil {
		return err
	}

	return nil
}

// +kubebuilder:rbac:groups=apps.adetalhouet.io,resources=ordersystems,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.adetalhouet.io,resources=ordersystems/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch

// Reconcile blah
func (reconciler *OrderSystemReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name)
	reqLogger.Info("Reconciling OrderSystem")

	// Fetch the CRD instance
	instance := &appsv1alpha1.OrderSystem{}
	err := reconciler.GetClient().Get(context.Background(), req.NamespacedName, instance)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("OrderSystem resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get OrderSystem")
		return reconciler.ManageError(instance, err)
	}

	// Managing CR validation
	if ok, err := reconciler.isValid(instance); !ok {
		return reconciler.ManageError(instance, err)
	}

	// Managing CR Initialization
	if ok := reconciler.isInitialized(instance); !ok {
		err := reconciler.GetClient().Update(context.Background(), instance)
		log.Info("Update object isInitialized")
		if err != nil {
			log.Error(err, "unable to update instance", "instance", instance)
			return reconciler.ManageError(instance, err)
		}
		return reconcile.Result{}, nil
	}

	// Managing CR Finalization
	if util.IsBeingDeleted(instance) {
		if !util.HasFinalizer(instance, orderSystemFinalizer) {
			return reconcile.Result{}, nil
		}
		err := reconciler.manageCleanUpLogic(instance)
		log.Info("Update object status manageCleanUpLogic")

		if err != nil {
			log.Error(err, "unable to delete instance", "instance", instance)
			return reconciler.ManageError(instance, err)
		}
		util.RemoveFinalizer(instance, orderSystemFinalizer)
		err = reconciler.GetClient().Update(context.Background(), instance)
		log.Info("Update object RemoveFinalizer")
		if err != nil {
			log.Error(err, "unable to update instance", "instance", instance)
			return reconciler.ManageError(instance, err)
		}
		return reconcile.Result{}, nil
	}

	// Managing Order System Logic
	err = reconciler.manageOperatorLogic(instance)
	if err != nil {
		return reconciler.ManageError(instance, err)
	}

	log.Info("Update object status end-of-loop")
	return reconciler.ManageSuccess(instance)
}

func (reconciler *OrderSystemReconciler) isInitialized(obj metav1.Object) bool {
	orderSystem, ok := obj.(*appsv1alpha1.OrderSystem)
	if !ok {
		return false
	}
	if util.HasFinalizer(orderSystem, orderSystemFinalizer) {
		return true
	}
	util.AddFinalizer(orderSystem, orderSystemFinalizer)
	return false

}

func (reconciler *OrderSystemReconciler) isValid(obj metav1.Object) (bool, error) {
	orderSystem, ok := obj.(*appsv1alpha1.OrderSystem)
	if !ok {
		return false, errors.New("not an OrderSystem object")
	}

	// Validate the nats and db services exist
	services := []string{orderSystem.Spec.DbInfo.Service, orderSystem.Spec.NatsInfo.Service}
	service := &corev1.Service{}
	err := reconciler.checkIfResourcesExist(orderSystem, services, service)
	if err != nil {
		return false, err
	}

	// Validate the nats and db secrets exists
	secrets := []string{orderSystem.Spec.DbInfo.Secret, orderSystem.Spec.NatsInfo.Secret}
	secret := &corev1.Secret{}
	err = reconciler.checkIfResourcesExist(orderSystem, secrets, secret)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (reconciler *OrderSystemReconciler) checkIfResourcesExist(orderSystem *appsv1alpha1.OrderSystem, objs []string, obj runtime.Object) error {
	for _, objName := range objs {
		objNamespaceName := types.NamespacedName{
			Name:      objName,
			Namespace: orderSystem.Namespace,
		}
		err := reconciler.GetClient().Get(context.Background(), objNamespaceName, obj)
		if err != nil {
			if apierrors.IsNotFound(err) {
				return errors.New("provided object(" + objNamespaceName.String() + ") doesn't exist")
			}
		}
	}
	return nil
}

func (reconciler *OrderSystemReconciler) manageCleanUpLogic(orderSystem *appsv1alpha1.OrderSystem) error {
	return reconciler.handleResources(orderSystem, true)
}

func (reconciler *OrderSystemReconciler) manageOperatorLogic(orderSystem *appsv1alpha1.OrderSystem) error {

	err := reconciler.handleResources(orderSystem, false)
	if err != nil {
		return err
	}

	// TODO handle istio
	// TODO handle autoscale

	return nil
}

func (reconciler *OrderSystemReconciler) handleResources(orderSystem *appsv1alpha1.OrderSystem, isDelete bool) error {
	resources := []apis.Resource{}

	var apps = map[string]templates.Service{
		"cart-service":    templates.Service{Port: 9090, Type: corev1.ServiceTypeClusterIP},
		"client-service":  templates.Service{Port: 9091, Type: corev1.ServiceTypeClusterIP},
		"order-service":   templates.Service{Port: 9092, Type: corev1.ServiceTypeClusterIP},
		"product-service": templates.Service{Port: 9093, Type: corev1.ServiceTypeClusterIP},
		"api-gw-service":  templates.Service{Port: 8080, Type: corev1.ServiceTypeLoadBalancer}}

	resources = append(resources, templates.ConfigMapSpec(orderSystem, apps))

	for podName, service := range apps {
		depName := templates.GetDeploymentName(orderSystem, podName)
		// TODO inject secret as ENV
		resources = append(resources, templates.DeploymentSpec(orderSystem, depName, podName, service))
		resources = append(resources, templates.ServiceSpec(orderSystem, depName, service))
	}

	if isDelete {
		return reconciler.DeleteResourcesIfExist(resources)
	}
	return reconciler.CreateResourcesIfNotExist(orderSystem, orderSystem.Namespace, resources)
}
