/*
Copyright 2024.

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

package controller

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	probev1 "github.com/mikebz/edge-status/api/v1"
)

// CheckReconciler reconciles a Check object
type CheckReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=probe.mikebz.com,resources=checks,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=probe.mikebz.com,resources=checks/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=probe.mikebz.com,resources=checks/finalizers,verbs=update

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *CheckReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	logv1 := log.V(1)

	var check probev1.Check
	if err := r.Get(ctx, req.NamespacedName, &check); err != nil {
		log.Error(err, "unable to fetch Check", "name", req.NamespacedName)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	configMapList := &corev1.ConfigMapList{}
	if err := r.List(ctx, configMapList, client.InNamespace(req.Namespace)); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to list configmaps: %w", err)
	}

	total := 0
	installed := false

	for _, configMap := range configMapList.Items {
		logv1.Info("Found ConfigMap", "name", configMap.Name)

		total++
		installed = true
	}

	check.Status.Total = total
	check.Status.Enabled = installed

	if err := r.Status().Update(ctx, &check); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to update check status: %w", err)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CheckReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&probev1.Check{}).
		Complete(r)
}
