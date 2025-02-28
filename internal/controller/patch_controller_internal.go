package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	patchesv1 "joshb.io/patchworks/api/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// RenderTemplate applies Go templating to the patch spec.
func RenderTemplate(tmpl string, data map[string]string) (string, error) {
	tpl, err := template.New("patch").Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("failed to parse: %w", err)
	}

	var rendered bytes.Buffer
	if err = tpl.Execute(&rendered, data); err != nil {
		return "", err
	}

	return rendered.String(), nil
}

func (r *PatchReconciler) getTargetResource(ctx context.Context, target patchesv1.TargetRef) (*unstructured.Unstructured, error) {
	gvk := schema.GroupVersionKind{
		Group:   extractGroup(target.APIVersion),
		Version: extractVersion(target.APIVersion),
		Kind:    target.Kind,
	}

	newTarget := &unstructured.Unstructured{}
	newTarget.SetGroupVersionKind(gvk)

	err := r.Get(ctx, client.ObjectKey{
		Namespace: target.Namespace,
		Name:      target.Name,
	}, newTarget)

	if err != nil {
		return nil, fmt.Errorf("failed to get target resource: %w", err)
	}

	return newTarget, nil
}

// extracts the data from the source kind
func (r *PatchReconciler) getDataFromSource(ctx context.Context, patchSpec patchesv1.PatchSpec) (map[string]string, error) {
	data := make(map[string]string)

	switch patchSpec.Source.Kind {
	case "ConfigMap":
		configMap := &v1.ConfigMap{}
		key := types.NamespacedName{Name: patchSpec.Source.Name, Namespace: patchSpec.Target.Namespace}
		if err := r.Get(ctx, key, configMap); err != nil {
			return nil, err
		}
		data = configMap.Data

	case "Secret":
		secret := &v1.Secret{}
		key := types.NamespacedName{Name: patchSpec.Source.Name, Namespace: patchSpec.Target.Namespace}
		if err := r.Get(ctx, key, secret); err != nil {
			return nil, err
		}
		for key, value := range secret.Data {
			data[key] = string(value) // Convert byte array to string
		}

	default:
		return nil, fmt.Errorf("unsupported source kind: %s", patchSpec.Source.Kind)
	}

	return data, nil

}

func (r *PatchReconciler) applyPatch(ctx context.Context, target patchesv1.TargetRef, yamlData string) error {
	// Convert rendered YAML into an unstructured object
	var patchObj map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlData), &patchObj); err != nil {
		return fmt.Errorf("failed to parse rendered YAML: %w", err)
	}

	// Convert target API version into GroupVersionKind (GVK)
	gv, err := schema.ParseGroupVersion(target.APIVersion)
	if err != nil {
		return fmt.Errorf("failed to parse API version: %w", err)
	}
	gvk := gv.WithKind(target.Kind)

	// Create an unstructured object to represent the target resource
	targetObj := &unstructured.Unstructured{}
	targetObj.SetGroupVersionKind(gvk)
	targetKey := types.NamespacedName{Name: target.Name, Namespace: target.Namespace}

	// Fetch the current version of the resource
	if err := r.Get(ctx, targetKey, targetObj); err != nil {
		return fmt.Errorf("failed to fetch target resource: %w", err)
	}

	originalState, _ := json.Marshal(targetObj.Object)
	annotations := map[string]string{
		"patches.joshb.io/patched-by":     "patch-operator",
		"patches.joshb.io/patch-id":       string(uuid.NewUUID()),
		"patches.joshb.io/original-state": string(originalState),
	}

	if err := addAnnotationsToUnstructured(targetObj, annotations); err != nil {
		return fmt.Errorf("failed to add annotations: %w", err)
	}

	// Convert the patch into JSON format
	patchBytes, err := json.Marshal(patchObj)
	if err != nil {
		return fmt.Errorf("failed to marshal patch JSON: %w", err)
	}

	// Apply the patch using a strategic merge patch
	if err := r.Patch(ctx, targetObj, client.RawPatch(types.StrategicMergePatchType, patchBytes)); err != nil {
		return fmt.Errorf("failed to apply patch: %w", err)
	}

	return nil
}

func (r *PatchReconciler) cleanup(ctx context.Context, patch *patchesv1.Patch) error {
	log := log.FromContext(ctx)
	log.Info("Performing cleanup", "Patch", patch.Name)

	//Reverse changes to target
	targetRef := patch.Spec.Target
	log.Info("Reverting patches", "Target", targetRef)

	var target unstructured.Unstructured
	target.SetAPIVersion(targetRef.APIVersion)
	target.SetKind(targetRef.Kind)

	if err := r.Get(ctx, client.ObjectKey{Name: targetRef.Name, Namespace: patch.Namespace}, &target); err != nil {
		log.Error(err, "Failed to retrieve target")
		return client.IgnoreNotFound(err)
	}

	// Retrieve original state
	originalStateJSON, exists := target.GetAnnotations()["patch.joshb.io/originalState"]
	if exists {
		var originalState map[string]interface{}
		if err := json.Unmarshal([]byte(originalStateJSON), &originalState); err == nil {
			target.Object = originalState
		}
	}

	// remove annotations
	annotations := target.GetAnnotations()
	log.Info("Reverting patch", "Patch", annotations["patches.joshb.io/patch-id"])
	delete(annotations, "patches.joshb.io/patch-id")
	delete(annotations, "patches.joshb.io/original-state")
	delete(annotations, "patches.joshb.io/patched-by")
	target.SetAnnotations(annotations)

	if err := r.Update(ctx, &target); err != nil {
		return err
	}

	log.Info("Cleanup complete", "Patch", patch.Name)
	return nil
}

func extractGroup(apiVersion string) string {
	parts := strings.Split(apiVersion, "/")
	if len(parts) == 1 {
		return ""
	}
	return parts[0]
}

func extractVersion(apiVersion string) string {
	parts := strings.Split(apiVersion, "/")
	return parts[len(parts)-1]
}

func addAnnotationsToUnstructured(obj *unstructured.Unstructured, annotations map[string]string) error {
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
