package utils

import (
	"context"
	"reflect"
	"testing"

	patchesv1alpha1 "github.com/joshbrgs/patchworks/api/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestExtractGroup(t *testing.T) {
	tests := []struct {
		name       string
		apiVersion string
		want       string
	}{
		{"standard group/version", "apps/v1", "apps"},
		{"core version (no group)", "v1", ""},
		{"custom group/version", "custom.io/v2alpha1", "custom.io"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExtractGroup(tt.apiVersion); got != tt.want {
				t.Errorf("ExtractGroup() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractVersion(t *testing.T) {
	tests := []struct {
		name       string
		apiVersion string
		want       string
	}{
		{"standard group/version", "apps/v1", "v1"},
		{"core version", "v1", "v1"},
		{"custom group/version", "custom.io/v2alpha1", "v2alpha1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExtractVersion(tt.apiVersion); got != tt.want {
				t.Errorf("ExtractVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddAnnotationsToUnstructured(t *testing.T) {
	obj := &unstructured.Unstructured{}
	annotations := map[string]string{"key1": "value1", "key2": "value2"}

	err := AddAnnotationsToUnstructured(obj, annotations)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := obj.GetAnnotations()
	if !reflect.DeepEqual(got, annotations) {
		t.Errorf("AddAnnotationsToUnstructured() = %v, want %v", got, annotations)
	}
}

func TestConvertToMapStringInterface(t *testing.T) {
	tests := []struct {
		name string
		in   interface{}
		want interface{}
	}{
		{
			"map with interface{} keys",
			map[interface{}]interface{}{"key": "value", 1: "number"},
			map[string]interface{}{"key": "value", "1": "number"},
		},
		{
			"nested maps",
			map[interface{}]interface{}{"outer": map[interface{}]interface{}{1: "inner"}},
			map[string]interface{}{"outer": map[string]interface{}{"1": "inner"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConvertToMapStringInterface(tt.in); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConvertToMapStringInterface() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetResource(t *testing.T) {
	scheme := runtime.NewScheme()
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	tests := []struct {
		name    string
		target  patchesv1alpha1.TargetRef
		setup   func() client.Client
		wantErr bool
	}{
		{
			"valid resource retrieval",
			patchesv1alpha1.TargetRef{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       "test-deploy",
				Namespace:  "default",
			},
			func() client.Client {
				obj := &unstructured.Unstructured{}
				obj.SetGroupVersionKind(schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"})
				obj.SetNamespace("default")
				obj.SetName("test-deploy")
				return fake.NewClientBuilder().WithScheme(scheme).WithObjects(obj).Build()
			},
			false,
		},
		{
			"resource not found",
			patchesv1alpha1.TargetRef{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       "missing-deploy",
				Namespace:  "default",
			},
			func() client.Client { return fakeClient },
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			c := tt.setup()
			_, err := GetResource(ctx, c, tt.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetResource() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
