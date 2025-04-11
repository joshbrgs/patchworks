package patch

import (
	"context"
	"fmt"

	patchesv1alpha1 "github.com/bigideaslearning/patchworks/api/v1alpha1"
	"github.com/bigideaslearning/patchworks/pkg/store"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type PatchCommand interface {
	GetDataFromSource(ctx context.Context, client client.Client, spec patchesv1alpha1.PatchSpec) (map[string]string, error)
	RenderTemplate(template string, data map[string]string) (string, error)
	ApplyAnnotationToOriginal(ctx context.Context, client client.Client, target patchesv1alpha1.TargetRef, annotations map[string]string) error
	ApplyPatch(ctx context.Context, client client.Client, target patchesv1alpha1.TargetRef, yaml string) error
}

type PatchCommandAbstract struct{}

func (e *PatchCommandAbstract) GetDataFromSource(ctx context.Context, client client.Client, spec patchesv1alpha1.PatchSpec) (map[string]string, error) {
	return getDataFromSource(ctx, client, spec)
}

func (e *PatchCommandAbstract) RenderTemplate(template string, data map[string]string) (string, error) {
	return RenderTemplate(template, data)
}

func (e *PatchCommandAbstract) ApplyAnnotationToOriginal(ctx context.Context, client client.Client, target patchesv1alpha1.TargetRef, annotations map[string]string) error {
	return applyAnnotationToOriginal(ctx, client, target, annotations)
}

func (e *PatchCommandAbstract) ApplyPatch(ctx context.Context, client client.Client, target patchesv1alpha1.TargetRef, yaml string) error {
	return applyPatch(ctx, client, target, yaml)
}

type PatchCommandConcrete struct {
	client   client.Client
	ctx      context.Context
	db       *store.RedisStore
	patch    *patchesv1alpha1.Patch
	commands PatchCommand
}

func NewPatchCommand(ctx context.Context, client client.Client, db *store.RedisStore, patch *patchesv1alpha1.Patch, commands PatchCommand) *PatchCommandConcrete {
	return &PatchCommandConcrete{ctx: ctx, client: client, db: db, patch: patch, commands: commands}
}

func (p *PatchCommandConcrete) Execute() error {
	log := log.FromContext(p.ctx)

	log.Info("Executing Patch Command")

	key, err := p.db.SetOriginal(p.patch.Spec.Target)
	if err != nil {
		return fmt.Errorf("failed to store original: %w", err)
	}

	// Add annotations
	annotations := make(map[string]string)
	annotations[patchedByAnnotation] = "patchworks-operator"
	annotations[patchedIdAnnotation] = key

	data, err := p.commands.GetDataFromSource(p.ctx, p.client, p.patch.Spec)
	if err != nil {
		return fmt.Errorf("failed to get source data: %w", err)
	}

	renderedYaml, err := p.commands.RenderTemplate(p.patch.Spec.Template, data)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	if err = p.commands.ApplyAnnotationToOriginal(p.ctx, p.client, p.patch.Spec.Target, annotations); err != nil {
		return fmt.Errorf("failed to patch original kind: %w", err)
	}

	return p.commands.ApplyPatch(p.ctx, p.client, p.patch.Spec.Target, renderedYaml)
}
