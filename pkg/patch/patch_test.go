package patch

import (
	"context"
	"testing"

	patchesv1alpha1 "github.com/joshbrgs/patchworks/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type MockPatchCommandAbstract struct{}

func (m *MockPatchCommandAbstract) GetDataFromSource(ctx context.Context, client client.Client, spec patchesv1alpha1.PatchSpec) (map[string]string, error) {
	return map[string]string{"mockKey": "mockValue"}, nil
}

func (m *MockPatchCommandAbstract) RenderTemplate(template string, data map[string]string) (string, error) {
	return "mock-rendered-yaml", nil
}

func (m *MockPatchCommandAbstract) ApplyAnnotationToOriginal(ctx context.Context, client client.Client, target patchesv1alpha1.TargetRef) error {
	return nil
}

func (m *MockPatchCommandAbstract) ApplyPatch(ctx context.Context, client client.Client, target patchesv1alpha1.TargetRef, yaml string) error {
	return nil
}

// MockClient is a simple mock for the Kubernetes client.
type MockClients struct {
	client.Client
}

func TestNewPatchCommand(t *testing.T) {
	ctx := context.TODO()
	mockClient := &MockClients{}
	mockPatch := &patchesv1alpha1.Patch{}
	mockExecutor := &MockPatchCommandAbstract{}

	pc := NewPatchCommand(ctx, mockClient, mockPatch, mockExecutor)

	assert.NotNil(t, pc)
	assert.Equal(t, mockClient, pc.client)
	assert.Equal(t, mockPatch, pc.patch)
	assert.Equal(t, mockExecutor, pc.commands)
}

func TestPatchCommand_Execute(t *testing.T) {
	ctx := context.TODO()
	mockClient := &MockClients{}
	mockPatch := &patchesv1alpha1.Patch{}
	mockExecutor := &MockPatchCommandAbstract{}

	pc := NewPatchCommand(ctx, mockClient, mockPatch, mockExecutor)

	err := pc.Execute()

	assert.NoError(t, err, "Expected Execute to return no error")
}
