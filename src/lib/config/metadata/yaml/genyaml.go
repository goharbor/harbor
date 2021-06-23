//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package main

import (
	"github.com/goharbor/harbor/src/lib/config/metadata"
	"os"
	"text/template"
)

const cfgTemplate = `  Configurations:
    type: object
    properties: {{ range .Items }}
      {{ .Name }}:
        type: {{ .Type }}
        description: {{ .Description }} 
        x-omitempty: true
        x-isnullable: true{{ end }}
`

const responseTemplate = `  ConfigurationResponse:
    type: object
    properties: {{ range .Items }}
      {{ .Name }}:
        $ref: '#/definitions/{{ .Type }}'
        description: {{ .Description }} {{ end }}
`

type document struct {
	Items []templateItem
}

type templateItem struct {
	Name        string
	Type        string
	Description string
}

func userCfgItems(isResponse bool) []templateItem {
	items := make([]templateItem, 0)
	for _, i := range metadata.ConfigList {
		if i.Scope == metadata.SystemScope {
			continue
		}
		// skip scan_all_policy
		if i.Name == "scan_all_policy" {
			continue
		}
		if _, ok := i.ItemType.(*metadata.PasswordType); isResponse && ok {
			continue
		}
		t := requestType(i.ItemType)
		if isResponse {
			t = toResponseType(t)
		}
		item := templateItem{
			Name:        i.Name,
			Type:        t,
			Description: i.Description,
		}
		items = append(items, item)
	}
	return items
}

type yamlFile struct {
	Name       string
	IsResponse bool
	TempName   string
}

var responseMap = map[string]string{
	"string":  "StringConfigItem",
	"boolean": "BoolConfigItem",
	"integer": "IntegerConfigItem",
}

// Used to generate swagger file for config response and configurations
func main() {
	l := []yamlFile{
		{"configurations.yml", false, cfgTemplate},
		{"configurationsResponse.yml", true, responseTemplate},
	}
	for _, file := range l {
		f, err := os.Create(file.Name)
		if err != nil {
			panic(err)
		}
		doc := document{
			Items: userCfgItems(file.IsResponse),
		}
		tmpl, err := template.New("test").Parse(file.TempName)
		if err != nil {
			panic(err)
		}
		err = tmpl.Execute(f, doc)
		if err != nil {
			panic(err)
		}
		f.Close()
	}
}

func requestType(item metadata.Type) string {
	switch item.(type) {
	case *metadata.StringType:
		return "string"
	case *metadata.BoolType:
		return "boolean"
	case *metadata.IntType, *metadata.PortType, *metadata.QuotaType, *metadata.LdapScopeType, *metadata.Int64Type:
		return "integer"
	}
	return "string"
}

func toResponseType(typeName string) string {
	return responseMap[typeName]
}
