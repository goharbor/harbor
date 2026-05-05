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

package transform

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/template"
)

// maxPayloadTransformSize is the maximum allowed byte size for a payload transform template.
const maxPayloadTransformSize = 4096

// Apply executes the given template against the JSON payload.
// If the template is empty, the payload is returned unchanged.
func Apply(templateString string, rawJSONPayload string) (string, error) {
	if templateString == "" {
		return rawJSONPayload, nil
	}
	if len(templateString) > maxPayloadTransformSize {
		return "", fmt.Errorf("payload_transform exceeds max size of %d bytes", maxPayloadTransformSize)
	}

	var eventData map[string]any
	if err := json.Unmarshal([]byte(rawJSONPayload), &eventData); err != nil {
		return "", fmt.Errorf("invalid payload JSON: %w", err)
	}

	compiledTemplate, err := template.New("payload_transform").
		Option("missingkey=error"). // reject references to fields that do not exist
		Parse(templateString)
	if err != nil {
		return "", fmt.Errorf("invalid payload_transform template: %w", err)
	}

	var renderedOutput bytes.Buffer
	if err := compiledTemplate.Execute(&renderedOutput, eventData); err != nil {
		return "", fmt.Errorf("failed to execute payload_transform: %w", err)
	}

	return renderedOutput.String(), nil
}

// Validate checks if template is valid
func Validate(templateString string) error {
	if templateString == "" {
		return nil
	}

	if len(templateString) > maxPayloadTransformSize {
		return fmt.Errorf("payload_transform exceeds max size of %d bytes", maxPayloadTransformSize)
	}

	_, err := template.New("payload_transform").
		Option("missingkey=error").
		Parse(templateString)
	return err
}
