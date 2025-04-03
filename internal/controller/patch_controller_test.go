/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

// import (
// 	"context"
// 	"reflect"
// 	"testing"
//
// 	patchesv1alpha1 "github.com/joshbrgs/patchworks/api/v1alpha1"
// 	v1 "k8s.io/api/core/v1"
// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// 	"k8s.io/apimachinery/pkg/runtime"
// 	"k8s.io/apimachinery/pkg/types"
// 	"k8s.io/client-go/tools/record"
// 	ctrl "sigs.k8s.io/controller-runtime"
// 	"sigs.k8s.io/controller-runtime/pkg/client"
// 	"sigs.k8s.io/controller-runtime/pkg/client/fake"
// )
//
// func TestPatchReconciler_Reconcile(t *testing.T) {
// 	testScheme := runtime.NewScheme()
// 	_ = v1.AddToScheme(testScheme)
//
// 	type fields struct {
// 		Client   client.Client
// 		Scheme   *runtime.Scheme
// 		Recorder record.EventRecorder
// 	}
// 	type args struct {
// 		ctx context.Context
// 		req ctrl.Request
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		want    ctrl.Result
// 		wantErr bool
// 	}{
// 		{
// 			name: "Patch resource not found",
// 			fields: fields{
// 				Client:   fake.NewClientBuilder().WithScheme(testScheme).Build(),
// 				Scheme:   testScheme,
// 				Recorder: record.NewFakeRecorder(10),
// 			},
// 			args: args{
// 				ctx: context.TODO(),
// 				req: ctrl.Request{NamespacedName: types.NamespacedName{Name: "nonexistent", Namespace: "default"}},
// 			},
// 			want:    ctrl.Result{},
// 			wantErr: false, // Expected behavior when resource is not found
// 		},
// 		{
// 			name: "Patch successfully applied",
// 			fields: fields{
// 				Client: func() client.Client {
// 					patch := &patchesv1alpha1.Patch{
// 						ObjectMeta: metav1.ObjectMeta{Name: "test-patch", Namespace: "default"},
// 						Spec:       patchesv1alpha1.PatchSpec{Target: patchesv1alpha1.TargetRef{Kind: "Pod", Name: "target-pod"}},
// 					}
// 					return fake.NewClientBuilder().WithScheme(testScheme).WithObjects(patch).Build()
// 				}(),
// 				Scheme:   testScheme,
// 				Recorder: record.NewFakeRecorder(10),
// 			},
// 			args: args{
// 				ctx: context.TODO(),
// 				req: ctrl.Request{NamespacedName: types.NamespacedName{Name: "test-patch", Namespace: "default"}},
// 			},
// 			want:    ctrl.Result{},
// 			wantErr: false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			r := &PatchReconciler{
// 				Client:   tt.fields.Client,
// 				Scheme:   tt.fields.Scheme,
// 				Recorder: tt.fields.Recorder,
// 			}
// 			got, err := r.Reconcile(tt.args.ctx, tt.args.req)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("PatchReconciler.Reconcile() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("PatchReconciler.Reconcile() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
//
// func TestPatchReconciler_cleanup(t *testing.T) {
// 	testScheme := runtime.NewScheme()
// 	_ = v1.AddToScheme(testScheme)
//
// 	type fields struct {
// 		Client   client.Client
// 		Scheme   *runtime.Scheme
// 		Recorder record.EventRecorder
// 	}
// 	type args struct {
// 		ctx   context.Context
// 		patch *patchesv1alpha1.Patch
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		wantErr bool
// 	}{
// 		{
// 			name: "Cleanup successfully removes patch annotations",
// 			fields: fields{
// 				Client: func() client.Client {
// 					patch := &patchesv1alpha1.Patch{
// 						ObjectMeta: metav1.ObjectMeta{
// 							Name:      "test-patch",
// 							Namespace: "default",
// 							Annotations: map[string]string{
// 								originalStateAnnotation: `{"metadata":{"name":"target"}}`,
// 								patchedByAnnotation:     "patch-controller",
// 								patchedIdAnnotation:     "12345",
// 							},
// 						},
// 					}
// 					return fake.NewClientBuilder().WithScheme(testScheme).WithObjects(patch).Build()
// 				}(),
// 				Scheme:   testScheme,
// 				Recorder: record.NewFakeRecorder(10),
// 			},
// 			args: args{
// 				ctx: context.TODO(),
// 				patch: &patchesv1alpha1.Patch{
// 					ObjectMeta: metav1.ObjectMeta{Name: "test-patch", Namespace: "default"},
// 				},
// 			},
// 			wantErr: false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			r := &PatchReconciler{
// 				Client:   tt.fields.Client,
// 				Scheme:   tt.fields.Scheme,
// 				Recorder: tt.fields.Recorder,
// 			}
// 			if err := r.cleanup(tt.args.ctx, tt.args.patch); (err != nil) != tt.wantErr {
// 				t.Errorf("PatchReconciler.cleanup() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }
