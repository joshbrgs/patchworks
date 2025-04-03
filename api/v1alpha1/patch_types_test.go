package v1alpha1

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// TestPatchJSONMarshalling ensures Patch struct marshals/unmarshals correctly
func TestPatchJSONMarshalling(t *testing.T) {
	patch := Patch{
		Spec: PatchSpec{
			Target: TargetRef{
				APIVersion: "v1",
				Kind:       "Pod",
				Name:       "mypod",
				Namespace:  "default",
			},
			Source: SourceRef{
				Kind: "ConfigMap",
				Name: "myconfig",
			},
			Template: "patch-template",
		},
		Status: PatchStatus{
			Applied: true,
			Message: "Successfully applied",
		},
	}

	data, err := json.Marshal(patch)
	assert.NoError(t, err)

	var decoded Patch
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, patch, decoded)
}

// TestSchemeRegistration ensures the Patch type is registered correctly
func TestSchemeRegistration(t *testing.T) {
	scheme := runtime.NewScheme()
	err := SchemeBuilder.AddToScheme(scheme)
	assert.NoError(t, err)

	gvk := schema.GroupVersionKind{Group: "patches.patchworks.io", Version: "v1alpha1", Kind: "Patch"}
	obj, err := scheme.New(gvk)
	assert.NoError(t, err)
	assert.NotNil(t, obj)
}

// Fuzz test for JSON deserialization
func FuzzPatchJSONUnmarshal(f *testing.F) {
	testJSON := `{"spec": {"target": {"apiVersion": "v1", "kind": "Pod", "name": "mypod", "namespace": "default"},
    "source": {"kind": "ConfigMap", "name": "myconfig"}, "template": "patch-template"},
    "status": {"applied": true, "message": "Successfully applied"}}`
	f.Add(testJSON)

	f.Fuzz(func(t *testing.T, input string) {
		var patch Patch
		_ = json.Unmarshal([]byte(input), &patch)
	})
}

// Fuzz test for invalid inputs
func FuzzPatchInvalidJSON(f *testing.F) {
	f.Add("{invalid-json}") // Corrupt JSON input

	f.Fuzz(func(t *testing.T, input string) {
		var patch Patch
		_ = json.Unmarshal([]byte(input), &patch)
	})
}
