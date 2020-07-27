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
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const controllerName = "ordersystem_crd"

var log = logf.Log.WithName(controllerName)

// Add creates a new MyCRD Controller and adds it to the Manager. The Manager will set fields on the Controller
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
	c, err := controller.New("ordersystem_crd_controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// isOrderSysteDeployment := predicate.Funcs{
	// 	UpdateFunc: func(e event.UpdateEvent) bool {
	// 		return templates.IsOrderSystemLabels(e.MetaNew.GetLabels())
	// 	},
	// 	CreateFunc: func(e event.CreateEvent) bool {
	// 		return templates.IsOrderSystemLabels(e.Meta.GetLabels())
	// 	},
	// 	DeleteFunc: func(e event.DeleteEvent) bool {
	// 		return templates.IsOrderSystemLabels(e.Meta.GetLabels())
	// 	},
	// }

	// Watch for changes to primary resource OrderSystem
	err = c.Watch(&source.Kind{Type: &appsv1alpha1.OrderSystem{}}, &handler.EnqueueRequestForObject{}, util.ResourceGenerationOrFinalizerChangedPredicate{})
	if err != nil {
		return err
	}

	// Watch for OrderSystem deployment OrderSystem
	// err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
	// 	IsController: true,
	// }, isOrderSysteDeployment)

	return nil
}

// OrderSystemReconciler reconciles a OrderSystem object
type OrderSystemReconciler struct {
	util.ReconcilerBase
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
	err := reconciler.GetClient().Get(context.TODO(), req.NamespacedName, instance)
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
		err := reconciler.GetClient().Update(context.TODO(), instance)
		if err != nil {
			log.Error(err, "unable to update instance", "instance", instance)
			return reconciler.ManageError(instance, err)
		}
		return reconcile.Result{}, nil
	}

	// Managing CR Finalization
	if util.IsBeingDeleted(instance) {
		if !util.HasFinalizer(instance, controllerName) {
			return reconcile.Result{}, nil
		}
		err := reconciler.manageCleanUpLogic(instance)
		if err != nil {
			log.Error(err, "unable to delete instance", "instance", instance)
			return reconciler.ManageError(instance, err)
		}
		util.RemoveFinalizer(instance, controllerName)
		err = reconciler.GetClient().Update(context.TODO(), instance)
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
	mycrd, ok := obj.(*appsv1alpha1.OrderSystem)
	if !ok {
		return false
	}
	if mycrd.Spec.Initialized {
		return true
	}
	util.AddFinalizer(mycrd, controllerName)
	mycrd.Spec.Initialized = true
	return false

}

func (reconciler *OrderSystemReconciler) isValid(obj metav1.Object) (bool, error) {
	mycrd, ok := obj.(*appsv1alpha1.OrderSystem)
	if !ok {
		return false, errors.New("not a OrderSystem object")
	}
	if mycrd.Spec.Valid {
		return true, nil
	}
	return false, errors.New("not valid because blah blah")
}

func (reconciler *OrderSystemReconciler) manageCleanUpLogic(orderSystem *appsv1alpha1.OrderSystem) error {
	return nil
}

func (reconciler *OrderSystemReconciler) manageOperatorLogic(orderSystem *appsv1alpha1.OrderSystem) error {
	if orderSystem.Spec.Error {
		return errors.New("error because blah blah")
	}
	// Vaidate CRD instance exist, else create
	err := reconciler.validateIfDeploymentExistOrCreate(orderSystem)

	// TODO check service
	// TODO check database -- move db service name to CRD
	// TODO check configmap

	return err
}

func (reconciler *OrderSystemReconciler) validateIfDeploymentExistOrCreate(orderSystem *appsv1alpha1.OrderSystem) error {
	// this is the list of expected deployment for the OrderSystem CRD, along with the service port
	var deployments = map[string]int32{
		"cart-service":        9090,
		"client-service":      9091,
		"order-service":       9092,
		"product-service":     9093,
		"order-system-api-gw": 8080}

	for deployment, port := range deployments {
		err := reconciler.doValidateIfExistOrCreate(orderSystem, deployment, port)
		if err != nil {
			return err
		}
	}
	return nil
}

func (reconciler *OrderSystemReconciler) doValidateIfExistOrCreate(orderSystem *appsv1alpha1.OrderSystem, deploymentName string, containerPort int32) error {
	ctx := context.Background()
	currentDep := &appsv1.Deployment{}
	depName := templates.GetDeploymentName(orderSystem, deploymentName)
	log.WithValues("Deployment.Namespace", orderSystem.Namespace, "Deployment.Name", depName)

	// reconciler.CreateIfNotExistTemplatedResources()

	err := reconciler.GetClient().Get(ctx, types.NamespacedName{Name: depName, Namespace: orderSystem.Namespace}, currentDep)
	if err != nil && apierrors.IsNotFound(err) {
		log.Info("Deployment not found")
		dep := templates.DeploymentSpec(orderSystem, deploymentName, containerPort)
		log.Info("Created new Deployment")
		err = reconciler.GetClient().Create(ctx, dep)
		if err != nil {
			log.Error(err, "Failed to create new Deployment")
			return err
		}
		// Deployment created successfully - return and requeue
		// Set OrderSystem instance as the owner and controller
		ctrl.SetControllerReference(orderSystem, dep, reconciler.GetScheme())
		return nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return err
	}
	log.Info("Deployment already exist")
	return nil
}
