package patch

import (
	"context"

	patchesv1alpha1 "github.com/bigideaslearning/patchworks/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const patchFinalizer = "patch.finalizers.patchworks.io"

type FinalizerManager struct {
	client client.Client
	ctx    context.Context
	patch  *patchesv1alpha1.Patch
}

func NewFinalizerManager(ctx context.Context, client client.Client, patch *patchesv1alpha1.Patch) *FinalizerManager {
	return &FinalizerManager{ctx: ctx, client: client, patch: patch}
}

func (f *FinalizerManager) EnsureFinalizer() error {
	if !controllerutil.ContainsFinalizer(f.patch, patchFinalizer) {
		controllerutil.AddFinalizer(f.patch, patchFinalizer)
		return f.client.Update(f.ctx, f.patch)
	}
	return nil
}

func (f *FinalizerManager) HandleDeletion(cleanupFunc func() error) error {
	if controllerutil.ContainsFinalizer(f.patch, patchFinalizer) {
		if err := cleanupFunc(); err != nil {
			return err
		}
		controllerutil.RemoveFinalizer(f.patch, patchFinalizer)
		return f.client.Update(f.ctx, f.patch)
	}
	return nil
}
