// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rbac

import (
	"reflect"
	"testing"
)

func TestResource_Subresource(t *testing.T) {
	type args struct {
		resources []Resource
	}
	tests := []struct {
		name string
		res  Resource
		args args
		want Resource
	}{
		{
			name: "subresource image",
			res:  Resource("/project/1"),
			args: args{
				resources: []Resource{"image"},
			},
			want: Resource("/project/1/image"),
		},
		{
			name: "subresource image build-history",
			res:  Resource("/project/1"),
			args: args{
				resources: []Resource{"image", "12", "build-history"},
			},
			want: Resource("/project/1/image/12/build-history"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.res.Subresource(tt.args.resources...); got != tt.want {
				t.Errorf("Resource.Subresource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResource_GetNamespace(t *testing.T) {
	tests := []struct {
		name    string
		res     Resource
		want    Namespace
		wantErr bool
	}{
		{
			name:    "project namespace",
			res:     Resource("/project/1"),
			want:    &projectNamespace{int64(1), false},
			wantErr: false,
		},
		{
			name:    "unknow namespace",
			res:     Resource("/unknow/1"),
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.res.GetNamespace()
			if (err != nil) != tt.wantErr {
				t.Errorf("Resource.GetNamespace() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Resource.GetNamespace() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResource_RelativeTo(t *testing.T) {
	type args struct {
		other Resource
	}
	tests := []struct {
		name    string
		res     Resource
		args    args
		want    Resource
		wantErr bool
	}{
		{
			name:    "/project/1/image",
			res:     Resource("/project/1/image"),
			args:    args{other: Resource("/project/1")},
			want:    Resource("image"),
			wantErr: false,
		},
		{
			name:    "/project/1",
			res:     Resource("/project/1"),
			args:    args{other: Resource("/project/1")},
			want:    Resource("."),
			wantErr: false,
		},
		{
			name:    "/project/1",
			res:     Resource("/project/1"),
			args:    args{other: Resource("/system")},
			want:    Resource(""),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.res.RelativeTo(tt.args.other)
			if (err != nil) != tt.wantErr {
				t.Errorf("Resource.RelativeTo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Resource.RelativeTo() = %v, want %v", got, tt.want)
			}
		})
	}
}
