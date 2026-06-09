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

package formats

func init() {
	registerFormats(CustomFormat, defaultFormatter)
}

const (
	// CustomFormat is the type for custom payload format.
	// It uses the same Default JSON structure, with an optional
	// Go template applied by the webhook job via the custom_payload field.
	CustomFormat = "Custom"
)
