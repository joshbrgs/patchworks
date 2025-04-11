package patch

import (
	"context"
	"testing"

	patchesv1alpha1 "github.com/bigideaslearning/patchworks/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestGetDataFromSource(t *testing.T) {
	ctx := context.TODO()

	// Create a fake Kubernetes client with test objects
	scheme := runtime.NewScheme()
	_ = v1.AddToScheme(scheme)

	// Create test ConfigMap and Secret
	testConfigMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: "test-config", Namespace: "default"},
		Data: map[string]string{
			"configKey": "configValue",
		},
	}
	testSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "test-secret", Namespace: "default"},
		Data: map[string][]byte{
			"secretKey": []byte("secretValue"),
		},
	}

	// Define test cases
	tests := []struct {
		name          string
		patchSpec     patchesv1alpha1.PatchSpec
		objects       []client.Object
		expectedData  map[string]string
		expectedError string
	}{
		{
			name: "Successfully fetch data from ConfigMap",
			patchSpec: patchesv1alpha1.PatchSpec{
				Source: patchesv1alpha1.SourceRef{
					Kind: "ConfigMap",
					Name: "test-config",
				},
				Target: patchesv1alpha1.TargetRef{Namespace: "default"},
			},
			objects:      []client.Object{testConfigMap},
			expectedData: map[string]string{"configKey": "configValue"},
		},
		{
			name: "Successfully fetch data from Secret",
			patchSpec: patchesv1alpha1.PatchSpec{
				Source: patchesv1alpha1.SourceRef{
					Kind: "Secret",
					Name: "test-secret",
				},
				Target: patchesv1alpha1.TargetRef{Namespace: "default"},
			},
			objects:      []client.Object{testSecret},
			expectedData: map[string]string{"secretKey": "secretValue"},
		},
		{
			name: "ConfigMap not found",
			patchSpec: patchesv1alpha1.PatchSpec{
				Source: patchesv1alpha1.SourceRef{
					Kind: "ConfigMap",
					Name: "missing-config",
				},
				Target: patchesv1alpha1.TargetRef{Namespace: "default"},
			},
			objects:       []client.Object{}, // No ConfigMap exists
			expectedError: "not found",
		},
		{
			name: "Secret not found",
			patchSpec: patchesv1alpha1.PatchSpec{
				Source: patchesv1alpha1.SourceRef{
					Kind: "Secret",
					Name: "missing-secret",
				},
				Target: patchesv1alpha1.TargetRef{Namespace: "default"},
			},
			objects:       []client.Object{}, // No Secret exists
			expectedError: "not found",
		},
		{
			name: "Unsupported source kind",
			patchSpec: patchesv1alpha1.PatchSpec{
				Source: patchesv1alpha1.SourceRef{
					Kind: "UnsupportedKind",
					Name: "some-name",
				},
				Target: patchesv1alpha1.TargetRef{Namespace: "default"},
			},
			expectedError: "unsupported source kind",
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(tt.objects...).Build()

			data, err := getDataFromSource(ctx, fakeClient, tt.patchSpec)

			if tt.expectedError != "" {
				assert.Error(t, err, "Expected an error but got nil")
				assert.Contains(t, err.Error(), tt.expectedError, "Error message mismatch")
			} else {
				assert.NoError(t, err, "Expected no error but got one")
				assert.Equal(t, tt.expectedData, data, "Data mismatch")
			}
		})
	}
}
