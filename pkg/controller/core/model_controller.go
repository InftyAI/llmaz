/*
Copyright 2024 The InftyAI Team.

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

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	coreapi "github.com/inftyai/llmaz/api/core/v1alpha1"
)

// OpenModelReconciler reconciles a Model object
type ModelReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Record record.EventRecorder
}

func NewModelReconciler(client client.Client, scheme *runtime.Scheme, record record.EventRecorder) *ModelReconciler {
	return &ModelReconciler{
		Client: client,
		Scheme: scheme,
		Record: record,
	}
}

//+kubebuilder:rbac:groups=llmaz.io,resources=openmodels,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=llmaz.io,resources=openmodels/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=llmaz.io,resources=openmodels/finalizers,verbs=update
//+kubebuilder:rbac:groups=manta.io,resources=torrents,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *ModelReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	model := &coreapi.OpenModel{}
	if err := r.Get(ctx, types.NamespacedName{Name: req.Name}, model); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger.V(10).Info("reconcile Model", "Model", klog.KObj(model))

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ModelReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&coreapi.OpenModel{}).
		Complete(r)
}
