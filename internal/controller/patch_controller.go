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

	patchesv1alpha1 "github.com/bigideaslearning/patchworks/api/v1alpha1"
	"github.com/bigideaslearning/patchworks/pkg/patch"
	"github.com/bigideaslearning/patchworks/pkg/store"
	"github.com/bigideaslearning/patchworks/pkg/utils"
	"github.com/redis/go-redis/v9"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// PatchReconciler reconciles a Patch object
type PatchReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
	DB       *redis.Client
}

const (
	patchedByAnnotation = "patch.patchworks.io/patched-by"
	patchedIdAnnotation = "patch.patchworks.io/patch-id"
)

// +kubebuilder:rbac:groups=patches.patchworks.io,resources=patches,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=patches.patchworks.io,resources=patches/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=patches.patchworks.io,resources=patches/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.4/pkg/reconcile
func (r *PatchReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	db := store.NewStore(ctx, r.Client, r.DB)

	log.Info("Reconciling Patch CRD", "Patch Name", req.Name)

	// Get Patch Kind
	patchKind := &patchesv1alpha1.Patch{}
	if err := r.Get(ctx, req.NamespacedName, patchKind); err != nil {
		log.Error(err, "Error in getting the patch kind")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.Info("Patch Kind retrieved")
	r.Recorder.Event(patchKind, v1.EventTypeNormal, "PatchProcessing", "Processing patch for target resource")

	// Check that the Patch is not being deleted
	finalizerManager := patch.NewFinalizerManager(ctx, r.Client, patchKind)

	if patchKind.ObjectMeta.DeletionTimestamp.IsZero() {
		log.Info("Patch Kind does not need deleted, adding finalizer to ensure resource cleanup later")
		if err := finalizerManager.EnsureFinalizer(); err != nil {
			return ctrl.Result{}, err
		}
	} else {
		log.Info("Cleaning up Kind resource")
		return ctrl.Result{}, finalizerManager.HandleDeletion(func() error { return r.cleanup(ctx, db, patchKind) })
	}

	// Execute Command to Patch
	patchCommand := patch.NewPatchCommand(ctx, r.Client, db, patchKind, &patch.PatchCommandAbstract{})

	if err := patchCommand.Execute(); err != nil {
		return ctrl.Result{}, err
	}

	// Update patch status
	patchKind.Status.Applied = true
	patchKind.Status.Message = "Patch applied successfully"
	if err := r.Status().Update(ctx, patchKind); err != nil {
		return ctrl.Result{}, err
	}

	r.Recorder.Event(patchKind, v1.EventTypeNormal, "PatchApplied", "Successfully applied patch")
	log.Info("Patch successfully applied", "Target", patchKind.Spec.Target)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PatchReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&patchesv1alpha1.Patch{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: 1}).
		Complete(r)
}

func (r *PatchReconciler) cleanup(ctx context.Context, db *store.RedisStore, patch *patchesv1alpha1.Patch) error {
	log := log.FromContext(ctx)
	log.Info("Performing cleanup", "Patch", patch.Name)

	targetRef := patch.Spec.Target
	log.Info("Reverting patches", "Target", targetRef)

	// Fetch the most recent version of the resource
	current, err := utils.GetResource(ctx, r.Client, targetRef)
	if err != nil {
		return err
	}

	annotations := current.GetAnnotations()

	// Retrieve original state from Redis
	patchId, ok := annotations[patchedIdAnnotation]
	if !ok {
		log.Info("Failed to find a patchId annotation")
		return nil
	}

	originalState, err := db.Get(patchId)
	if err != nil {
		log.Error(err, "Failed to retrieve original", "Redis", patchId)
		return err
	}

	// Carry over updated metadata from `current` to `originalState`
	originalState.SetResourceVersion(current.GetResourceVersion())
	originalState.SetUID(current.GetUID())

	// Copy over any annotations from the current that shouldn't be reverted
	delete(annotations, patchedIdAnnotation)
	delete(annotations, patchedByAnnotation)

	originalState.SetAnnotations(annotations)

	// Perform update
	if err := r.Update(ctx, originalState); err != nil {
		log.Error(err, "Failed to update resource with original state")
		return err
	}

	log.Info("Cleanup complete", "Patch", patch.Name)
	r.Recorder.Event(patch, v1.EventTypeWarning, "PatchReverted", "Patch has been removed")
	return nil
}
