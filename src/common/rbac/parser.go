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
	"errors"
	"regexp"
	"strconv"
)

var (
	namespaceParsers = map[string]namespaceParser{
		"project": projectNamespaceParser,
	}
)

type namespaceParser func(resource Resource) (Namespace, error)

func projectNamespaceParser(resource Resource) (Namespace, error) {
	parserRe := regexp.MustCompile("^/project/([^/]*)/?")

	matches := parserRe.FindStringSubmatch(resource.String())

	if len(matches) <= 1 {
		return nil, errors.New("not support resource")
	}

	projectID, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return nil, err
	}

	return &projectNamespace{projectID: projectID}, nil
}
