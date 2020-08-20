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

package project

import (
	"context"
	"fmt"

	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/project/models"
)

// Result the result for ListAll func
type Result struct {
	Data  *models.Project
	Error error
}

// ListAll returns all projects with chunk support
func ListAll(ctx context.Context, chunkSize int, query *q.Query, options ...Option) <-chan Result {
	ch := make(chan Result, chunkSize)

	go func() {
		defer close(ch)

		query = q.MustClone(query)
		query.PageNumber = 1
		query.PageSize = int64(chunkSize)

		for {
			projects, err := Ctl.List(ctx, query, options...)
			if err != nil {
				format := "failed to list projects at page %d with page size %d, error :%v"
				ch <- Result{Error: fmt.Errorf(format, query.PageNumber, query.PageSize, err)}
				return
			}

			for _, p := range projects {
				ch <- Result{Data: p}
			}

			if len(projects) < chunkSize {
				break
			}

			query.PageNumber++
		}

	}()

	return ch
}
