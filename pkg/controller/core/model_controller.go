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

	manta "github.com/inftyai/manta/api/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
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

	if preheatEnabled(model) {
		torrent := manta.Torrent{}
		if err := r.Client.Get(ctx, types.NamespacedName{Name: model.Name}, &torrent); err != nil {
			if apierrors.IsNotFound(err) {
				logger.Info("preheating model", "model", klog.KObj(model))

				torrent := constructTorrent(model)
				if err := r.Client.Create(ctx, &torrent); err != nil {
					return ctrl.Result{}, err
				}
			} else {
				return ctrl.Result{}, err
			}
		}

		if setCondition(model, &torrent) {
			if err := r.Status().Update(ctx, model); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{}, nil
}

func setCondition(model *coreapi.OpenModel, torrent *manta.Torrent) (changed bool) {
	if len(model.Status.Conditions) == 0 {
		condition := metav1.Condition{
			Type:    coreapi.ModelPending,
			Status:  metav1.ConditionTrue,
			Reason:  "Pending",
			Message: "Waiting for model downloading",
		}
		return apimeta.SetStatusCondition(&model.Status.Conditions, condition)
	}

	if apimeta.IsStatusConditionTrue(torrent.Status.Conditions, manta.ReadyConditionType) {
		condition := metav1.Condition{
			Type:    coreapi.ModelReady,
			Status:  metav1.ConditionTrue,
			Reason:  "Ready",
			Message: "Model is downloaded successfully",
		}
		return apimeta.SetStatusCondition(&model.Status.Conditions, condition)
	}
	return false
}

// SetupWithManager sets up the controller with the Manager.
func (r *ModelReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&coreapi.OpenModel{}).
		Complete(r)
}

func preheatEnabled(model *coreapi.OpenModel) bool {
	return model.Annotations != nil && model.Annotations[coreapi.ModelPreheatAnnoKey] == "true" && model.Spec.Source.ModelHub != nil && *model.Spec.Source.ModelHub.Name == coreapi.HUGGING_FACE
}

func constructTorrent(model *coreapi.OpenModel) manta.Torrent {
	return manta.Torrent{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Torrent",
			APIVersion: manta.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: model.Name,
			OwnerReferences: []metav1.OwnerReference{
				{
					Kind:               "OpenModel",
					APIVersion:         coreapi.GroupVersion.String(),
					Name:               model.Name,
					UID:                model.UID,
					BlockOwnerDeletion: ptr.To(true),
					Controller:         ptr.To(true),
				},
			},
		},
		Spec: manta.TorrentSpec{
			Replicas:      ptr.To[int32](1),
			ReclaimPolicy: ptr.To(manta.DeleteReclaimPolicy),
			Hub: &manta.Hub{
				RepoID:   model.Spec.Source.ModelHub.ModelID,
				Filename: model.Spec.Source.ModelHub.Filename,
				Revision: model.Spec.Source.ModelHub.Revision,
			},
		},
	}
}
