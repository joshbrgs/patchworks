/*
Copyright 2025.

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

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	patchesv1 "joshb.io/patchworks/api/v1"
)

// PatchReconciler reconciles a Patch object
type PatchReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

const patchFinalizer = "patch.finalizers.joshb.io"

// +kubebuilder:rbac:groups=patches.joshb.io,resources=patches,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=patches.joshb.io,resources=patches/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=patches.joshb.io,resources=patches/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Patch object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.4/pkg/reconcile
func (r *PatchReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Reconciling Patch CRD", "Patch Name", req.Name)

	patch := &patchesv1.Patch{}
	if err := r.Get(ctx, req.NamespacedName, patch); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Check if the resource is being deleted
	if patch.ObjectMeta.DeletionTimestamp.IsZero() {
		// Ensure finalizer is added
		if !controllerutil.ContainsFinalizer(patch, patchFinalizer) {
			controllerutil.AddFinalizer(patch, patchFinalizer)
			if err := r.Update(ctx, patch); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		if controllerutil.ContainsFinalizer(patch, patchFinalizer) {
			// Perform cleanup before deleting the resource
			if err := r.cleanup(ctx, patch); err != nil {
				return ctrl.Result{}, err
			}

			// Remove the finalizer after cleanup
			controllerutil.RemoveFinalizer(patch, patchFinalizer)
			if err := r.Update(ctx, patch); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	r.Recorder.Event(patch, v1.EventTypeNormal, "PatchProcessing", "Processing patch for target resource")

	data, err := r.getDataFromSource(ctx, patch.Spec)
	if err != nil {
		log.Error(err, "Failed to render template")
		r.Recorder.Event(patch, v1.EventTypeWarning, "TemplateRenderFailed", "Failed to render template")
		return ctrl.Result{}, err
	}

	// Render template before patching
	renderedYaml, err := RenderTemplate(patch.Spec.Template, data)
	if err != nil {
		log.Error(err, "Failed to render template")
		r.Recorder.Event(patch, v1.EventTypeWarning, "TemplateRenderFailed", "Failed to render template")
		return ctrl.Result{}, err
	}

	log.Info("Final patched yaml", "yaml", renderedYaml)

	if err = r.applyPatch(ctx, patch.Spec.Target, renderedYaml); err != nil {
		log.Error(err, "Failed to apply patch")
		r.Recorder.Event(patch, v1.EventTypeWarning, "PatchFailed", "Failed to apply patch")
		return ctrl.Result{}, err
	}

	r.Recorder.Event(patch, v1.EventTypeNormal, "PatchApplied", "Successfully applied patch")
	log.Info("Patch successfully applied", "Target", patch.Spec.Target)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PatchReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&patchesv1.Patch{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: 1}).
		Complete(r)
}
