package patch

import (
	"context"
	"fmt"

	patchesv1alpha1 "github.com/joshbrgs/patchworks/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type PatchCommand interface {
	GetDataFromSource(ctx context.Context, client client.Client, spec patchesv1alpha1.PatchSpec) (map[string]string, error)
	RenderTemplate(template string, data map[string]string) (string, error)
	ApplyAnnotationToOriginal(ctx context.Context, client client.Client, target patchesv1alpha1.TargetRef) error
	ApplyPatch(ctx context.Context, client client.Client, target patchesv1alpha1.TargetRef, yaml string) error
}

type PatchCommandAbstract struct{}

func (e *PatchCommandAbstract) GetDataFromSource(ctx context.Context, client client.Client, spec patchesv1alpha1.PatchSpec) (map[string]string, error) {
	return getDataFromSource(ctx, client, spec)
}

func (e *PatchCommandAbstract) RenderTemplate(template string, data map[string]string) (string, error) {
	return RenderTemplate(template, data)
}

func (e *PatchCommandAbstract) ApplyAnnotationToOriginal(ctx context.Context, client client.Client, target patchesv1alpha1.TargetRef) error {
	return applyAnnotationToOriginal(ctx, client, target)
}

func (e *PatchCommandAbstract) ApplyPatch(ctx context.Context, client client.Client, target patchesv1alpha1.TargetRef, yaml string) error {
	return applyPatch(ctx, client, target, yaml)
}

type PatchCommandConcrete struct {
	client   client.Client
	ctx      context.Context
	patch    *patchesv1alpha1.Patch
	commands PatchCommand
}

func NewPatchCommand(ctx context.Context, client client.Client, patch *patchesv1alpha1.Patch, commands PatchCommand) *PatchCommandConcrete {
	return &PatchCommandConcrete{ctx: ctx, client: client, patch: patch, commands: commands}
}

func (p *PatchCommandConcrete) Execute() error {
	log := log.FromContext(p.ctx)

	log.Info("Executing Patch Command")

	data, err := p.commands.GetDataFromSource(p.ctx, p.client, p.patch.Spec)
	if err != nil {
		return fmt.Errorf("failed to get source data: %w", err)
	}

	renderedYaml, err := p.commands.RenderTemplate(p.patch.Spec.Template, data)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	if err = p.commands.ApplyAnnotationToOriginal(p.ctx, p.client, p.patch.Spec.Target); err != nil {
		return fmt.Errorf("failed to patch original kind: %w", err)
	}

	return p.commands.ApplyPatch(p.ctx, p.client, p.patch.Spec.Target, renderedYaml)
}
