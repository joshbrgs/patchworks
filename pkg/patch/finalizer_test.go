package patch

import (
	"context"
	"errors"
	"testing"

	patchesv1alpha1 "github.com/bigideaslearning/patchworks/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type MockClient struct {
	mock.Mock
	client.Client
}

func (m *MockClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	args := m.Called(ctx, obj)
	return args.Error(0)
}

func TestNewFinalizerManager(t *testing.T) {
	ctx := context.TODO()
	mockClient := new(MockClient)
	patch := &patchesv1alpha1.Patch{}

	fm := NewFinalizerManager(ctx, mockClient, patch)

	assert.NotNil(t, fm)
	assert.Equal(t, mockClient, fm.client)
	assert.Equal(t, ctx, fm.ctx)
	assert.Equal(t, patch, fm.patch)
}

func TestFinalizerManager_EnsureFinalizer(t *testing.T) {
	ctx := context.TODO()
	mockClient := new(MockClient)
	patch := &patchesv1alpha1.Patch{}

	fm := NewFinalizerManager(ctx, mockClient, patch)

	// Case: Finalizer does not exist, should add it and update client
	mockClient.On("Update", ctx, patch).Return(nil).Once()

	err := fm.EnsureFinalizer()
	assert.NoError(t, err)
	assert.Contains(t, patch.Finalizers, patchFinalizer)
	mockClient.AssertExpectations(t)
}

func TestFinalizerManager_HandleDeletion(t *testing.T) {
	ctx := context.TODO()
	mockClient := new(MockClient)
	patch := &patchesv1alpha1.Patch{}
	controllerutil.AddFinalizer(patch, patchFinalizer)

	fm := NewFinalizerManager(ctx, mockClient, patch)

	// Case: Cleanup function succeeds, finalizer should be removed
	mockClient.On("Update", ctx, patch).Return(nil).Once()
	cleanupFunc := func() error { return nil }

	err := fm.HandleDeletion(cleanupFunc)
	assert.NoError(t, err)
	assert.NotContains(t, patch.Finalizers, patchFinalizer)
	mockClient.AssertExpectations(t)

	// Case: Cleanup function fails, finalizer should remain
	controllerutil.AddFinalizer(patch, patchFinalizer)
	errorCleanup := errors.New("cleanup failed")
	cleanupFunc = func() error { return errorCleanup }

	err = fm.HandleDeletion(cleanupFunc)
	assert.Error(t, err)
	assert.Contains(t, patch.Finalizers, patchFinalizer)
}
