package patch

import (
	"context"
	"testing"

	patchesv1alpha1 "github.com/joshbrgs/patchworks/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Test_applyPatch(t *testing.T) {
	type args struct {
		ctx      context.Context
		c        client.Client
		target   patchesv1alpha1.TargetRef
		yamlData string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := applyPatch(tt.args.ctx, tt.args.c, tt.args.target, tt.args.yamlData); (err != nil) != tt.wantErr {
				t.Errorf("applyPatch() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_applyAnnotationToOriginal(t *testing.T) {
	type args struct {
		ctx    context.Context
		c      client.Client
		target patchesv1alpha1.TargetRef
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := applyAnnotationToOriginal(tt.args.ctx, tt.args.c, tt.args.target); (err != nil) != tt.wantErr {
				t.Errorf("applyAnnotationToOriginal() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
