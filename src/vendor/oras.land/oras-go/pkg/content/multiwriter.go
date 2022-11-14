/*
Copyright The ORAS Authors.
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

package content

import (
	"context"

	ctrcontent "github.com/containerd/containerd/content"
	"github.com/containerd/containerd/remotes"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// MultiWriterIngester an ingester that can provide a single writer or multiple writers for a single
// descriptor. Useful when the target of a descriptor can have multiple items within it, e.g. a layer
// that is a tar file with multiple files, each of which should go to a different stream, some of which
// should not be handled at all.
type MultiWriterIngester interface {
	ctrcontent.Ingester
	Writers(ctx context.Context, opts ...ctrcontent.WriterOpt) (func(string) (ctrcontent.Writer, error), error)
}

// MultiWriterPusher a pusher that can provide a single writer or multiple writers for a single
// descriptor. Useful when the target of a descriptor can have multiple items within it, e.g. a layer
// that is a tar file with multiple files, each of which should go to a different stream, some of which
// should not be handled at all.
type MultiWriterPusher interface {
	remotes.Pusher
	Pushers(ctx context.Context, desc ocispec.Descriptor) (func(string) (ctrcontent.Writer, error), error)
}
