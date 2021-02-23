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

package artifact

import (
	"context"

	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
)

// Iterator returns the iterator to fetch all artifacts with query
func Iterator(ctx context.Context, chunkSize int, query *q.Query, option *Option) <-chan *Artifact {
	ch := make(chan *Artifact, chunkSize)

	go func() {
		defer close(ch)

		clone := q.MustClone(query)
		clone.PageNumber = 1
		clone.PageSize = int64(chunkSize)

		for {
			artifacts, err := Ctl.List(ctx, clone, option)
			if err != nil {
				log.G(ctx).Errorf("list artifacts failed, error: %v", err)
				return
			}

			for _, artifact := range artifacts {
				ch <- artifact
			}

			if len(artifacts) < chunkSize {
				break
			}

			clone.PageNumber++
		}

	}()

	return ch
}
