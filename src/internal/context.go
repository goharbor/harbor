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

package internal

import "context"

type contextKey string

// define all context key here to avoid conflict
const (
	contextKeyAPIVersion contextKey = "apiVersion"
)

func setToContext(ctx context.Context, key contextKey, value interface{}) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, key, value)
}

func getFromContext(ctx context.Context, key contextKey) interface{} {
	if ctx == nil {
		return nil
	}
	return ctx.Value(key)
}

// SetAPIVersion sets the API version into the context
func SetAPIVersion(ctx context.Context, version string) context.Context {
	return setToContext(ctx, contextKeyAPIVersion, version)
}

// GetAPIVersion gets the API version from the context
func GetAPIVersion(ctx context.Context) string {
	version := ""
	value := getFromContext(ctx, contextKeyAPIVersion)
	if value != nil {
		version = value.(string)
	}
	return version
}
