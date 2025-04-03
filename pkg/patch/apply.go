package patch

import (
	"context"
	"encoding/json"
	"fmt"

	patchesv1alpha1 "github.com/joshbrgs/patchworks/api/v1alpha1"
	"github.com/joshbrgs/patchworks/pkg/utils"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/uuid"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	originalStateAnnotation = "patch.patchworks.io/original-state"
	patchedByAnnotation     = "patch.patchworks.io/patched-by"
	patchedIdAnnotation     = "patch.patchworks.io/patch-id"
)

// Applies patch based on the target supplied and yamlData
func applyPatch(ctx context.Context, c client.Client, target patchesv1alpha1.TargetRef, yamlData string) error {
	log := log.FromContext(ctx)

	targetObj, err := utils.GetResource(ctx, c, target)
	if err != nil {
		return fmt.Errorf("failed to get resource: %w", err)
	}

	// Convert rendered YAML into an unstructured object
	var rawPatchObj map[interface{}]interface{}
	if err := yaml.Unmarshal([]byte(yamlData), &rawPatchObj); err != nil {
		return fmt.Errorf("failed to parse rendered YAML: %w", err)
	}

	// Convert to map[string]interface{}
	patchObj := utils.ConvertToMapStringInterface(rawPatchObj)
	log.Info("PatchObj", "Patch", patchObj)

	// Convert the patch into JSON format
	patchBytes, err := json.Marshal(patchObj)
	if err != nil {
		return fmt.Errorf("failed to marshal patch JSON: %w", err)
	}

	// Apply the patch using a strategic merge patch
	if err := c.Patch(ctx, targetObj, client.RawPatch(types.StrategicMergePatchType, patchBytes)); err != nil {
		return fmt.Errorf("failed to apply patch: %w", err)
	}

	return nil
}

func applyAnnotationToOriginal(ctx context.Context, c client.Client, target patchesv1alpha1.TargetRef) error {

	// Get the object of the target to patch
	targetObj, err := utils.GetResource(ctx, c, target)
	if err != nil {
		return fmt.Errorf("failed to get resource: %w", err)
	}

	// originalState, _ := json.Marshal(targetObj.Object)

	// Add annotations
	annotations := make(map[string]string)
	// annotations[originalStateAnnotation] = string(originalState)
	annotations[patchedByAnnotation] = "patchworks-operator"
	annotations[patchedIdAnnotation] = string(uuid.NewUUID())

	if err := utils.AddAnnotationsToUnstructured(targetObj, annotations); err != nil {
		return fmt.Errorf("failed to add annotations: %w", err)
	}

	// Persist the annotations update separately before patching
	if err := c.Update(ctx, targetObj); err != nil {
		return fmt.Errorf("failed to update target object with annotations: %w", err)
	}

	return nil
}
