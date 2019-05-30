// Copyright 2019 Project Harbor Authors
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

package chain

import (
	"fmt"

	"github.com/goharbor/harbor/src/common/retention"
	"github.com/goharbor/harbor/src/retention/filter"
)

// Build constructs an ordered slice of retention filters for the given retention policy
func Build(p *retention.Policy) (filters []retention.Filter, err error) {
	for _, meta := range p.Filters {
		var impl retention.Filter

		switch meta.Type {
		case filter.TypeDeleteOlderThan:
			impl, err = filter.NewDeleteOlderThan(meta.Options)
		case filter.TypeDeleteRegex:
			impl, err = filter.NewDeleteRegex(meta.Options)
		case filter.TypeKeepEverything:
			impl = &filter.KeepEverything{}
		case filter.TypeKeepMostRecentN:
			impl, err = filter.NewKeepMostRecentN(meta.Options)
		case filter.TypeKeepRegex:
			impl, err = filter.NewKeepRegex(meta.Options)
		case filter.TypeDeleteEverything:
			impl = &filter.DeleteEverything{}
		default:
			err = fmt.Errorf("unknown filter type: %s", meta.Type)
		}

		if err != nil {
			return
		}

		filters = append(filters, impl)
	}

	return
}
