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
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	return nil
}

// +kubebuilder:rbac:groups=apps.adetalhouet.io,resources=ordersystems,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.adetalhouet.io,resources=ordersystems/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;

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
		if err != nil {
			log.Error(err, "unable to delete instance", "instance", instance)
			return reconciler.ManageError(instance, err)
		}
		util.RemoveFinalizer(instance, orderSystemFinalizer)
		err = reconciler.GetClient().Update(context.Background(), instance)
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
	_, ok := obj.(*appsv1alpha1.OrderSystem)
	if !ok {
		return false, errors.New("not an OrderSystem object")
	}
	return true, nil
}

func (reconciler *OrderSystemReconciler) manageCleanUpLogic(orderSystem *appsv1alpha1.OrderSystem) error {
	return reconciler.handleDeployment(orderSystem, true)
}

func (reconciler *OrderSystemReconciler) manageOperatorLogic(orderSystem *appsv1alpha1.OrderSystem) error {
	err := reconciler.handleDeployment(orderSystem, false)
	if err != nil {
		return err
	}

	// TODO check service
	// TODO check database -- move db service name to CRD
	// TODO check configmap

	return nil
}

func (reconciler *OrderSystemReconciler) handleDeployment(orderSystem *appsv1alpha1.OrderSystem, isDelete bool) error {
	// this is the list of expected deployment for the OrderSystem CRD, along with the service port
	var deployments = map[string]int32{
		"cart-service":        9090,
		"client-service":      9091,
		"order-service":       9092,
		"product-service":     9093,
		"order-system-api-gw": 8080}

	for name, port := range deployments {

		depName := templates.GetDeploymentName(orderSystem, name)
		depLog := log.WithValues("Deployment.Namespace", orderSystem.Namespace, "Deployment.Name", depName)
		dep := templates.DeploymentSpec(orderSystem, name, port)

		if !isDelete {
			depLog.Info("Create deployment if not exists")
			err := reconciler.CreateResourceIfNotExists(orderSystem, orderSystem.Namespace, dep)
			if err != nil {
				depLog.Error(err, "fail to create instance", "instance", orderSystem)
				return err
			}
		} else {
			depLog.Info("Delete deployment if exists")
			err := reconciler.DeleteResourceIfExists(dep)
			if err != nil {
				depLog.Error(err, "fail to delete instance", "instance", orderSystem)
				return err
			}
		}
	}
	return nil
}
