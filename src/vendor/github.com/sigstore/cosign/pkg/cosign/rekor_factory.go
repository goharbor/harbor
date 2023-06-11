//
// Copyright 2022 The Sigstore Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cosign

import (
	"context"

	"github.com/sigstore/rekor/pkg/generated/client"
)

// key is used for associating the Rekor client client inside the
// context.Context.
type key struct{}

// TODO(jason): Rename this to something better than pkg/cosign.Set.
func Set(ctx context.Context, rekorClient *client.Rekor) context.Context {
	return context.WithValue(ctx, key{}, rekorClient)
}

// Get extracts the Rekor client from the context.
// TODO(jason): Rename this to something better than pkg/cosign.Get.
func Get(ctx context.Context) *client.Rekor {
	untyped := ctx.Value(key{})
	if untyped == nil {
		return nil
	}
	return untyped.(*client.Rekor)
}
