package patch

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/log"

	patchesv1alpha1 "github.com/joshbrgs/patchworks/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getDataFromSource(ctx context.Context, c client.Client, patchSpec patchesv1alpha1.PatchSpec) (map[string]string, error) {
	data := make(map[string]string)
	log := log.FromContext(ctx)

	log.Info("getting source data from kind", "Kind", patchSpec.Source.Kind, "Name", patchSpec.Source.Name)

	switch patchSpec.Source.Kind {
	case "ConfigMap":
		configMap := &v1.ConfigMap{}
		key := types.NamespacedName{Name: patchSpec.Source.Name, Namespace: patchSpec.Target.Namespace}

		if err := c.Get(ctx, key, configMap); err != nil {
			return nil, err
		}

		log.Info("retrieved source data from", "ConfigMap", configMap.Data)

		data = configMap.Data

	case "Secret":
		secret := &v1.Secret{}
		key := types.NamespacedName{Name: patchSpec.Source.Name, Namespace: patchSpec.Target.Namespace}

		if err := c.Get(ctx, key, secret); err != nil {
			return nil, err
		}

		for key, value := range secret.Data {
			data[key] = string(value) // Convert byte array to string
		}
		log.Info("retrieved source data from", "Secret", data)

	default:
		return nil, fmt.Errorf("unsupported source kind: %s", patchSpec.Source.Kind)
	}

	return data, nil
}
