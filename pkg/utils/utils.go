package utils

import (
	"context"
	"fmt"
	"strings"

	patchesv1alpha1 "github.com/joshbrgs/patchworks/api/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ExtractGroup(apiVersion string) string {
	parts := strings.Split(apiVersion, "/")
	if len(parts) == 1 {
		return ""
	}
	return parts[0]
}

func ExtractVersion(apiVersion string) string {
	parts := strings.Split(apiVersion, "/")
	return parts[len(parts)-1]
}

func AddAnnotationsToUnstructured(obj *unstructured.Unstructured, annotations map[string]string) error {
	// Retrieve existing annotations (if any)
	existingAnnotations := obj.GetAnnotations()
	if existingAnnotations == nil {
		existingAnnotations = make(map[string]string)
	}

	// Add or update annotations
	for key, value := range annotations {
		existingAnnotations[key] = value
	}

	// Set the updated annotations
	obj.SetAnnotations(existingAnnotations)

	return nil
}

func GetResource(ctx context.Context, c client.Client, target patchesv1alpha1.TargetRef) (*unstructured.Unstructured, error) {
	gvk := schema.GroupVersionKind{
		Group:   ExtractGroup(target.APIVersion),
		Version: ExtractVersion(target.APIVersion),
		Kind:    target.Kind,
	}

	if gvk.Empty() {
		return nil, fmt.Errorf("failed to get target resource with the provided group, version, or kind")
	}

	newTarget := &unstructured.Unstructured{}
	newTarget.SetGroupVersionKind(gvk)

	targetKey := types.NamespacedName{
		Namespace: target.Namespace,
		Name:      target.Name,
	}

	err := c.Get(ctx, targetKey, newTarget)

	if err != nil {
		return nil, fmt.Errorf("failed to get target resource: %w", err)
	}

	return newTarget, nil
}

func ConvertToMapStringInterface(in interface{}) interface{} {
	switch v := in.(type) {
	case map[interface{}]interface{}:
		out := make(map[string]interface{})
		for key, value := range v {
			strKey, ok := key.(string)
			if !ok {
				strKey = fmt.Sprint(key) // Convert non-string keys to string
			}
			out[strKey] = ConvertToMapStringInterface(value)
		}
		return out
	case []interface{}:
		for i, elem := range v {
			v[i] = ConvertToMapStringInterface(elem)
		}
	}
	return in
}
