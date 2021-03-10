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

package ldap

import (
	"strings"

	ber "github.com/go-asn1-ber/asn1-ber"
	goldap "github.com/go-ldap/ldap/v3"
)

// FilterBuilder build filter for ldap search
type FilterBuilder struct {
	packet *ber.Packet
}

// Or ...
func (f *FilterBuilder) Or(filterB *FilterBuilder) *FilterBuilder {
	if f.packet == nil {
		return filterB
	}
	if filterB.packet == nil {
		return f
	}
	p := ber.Encode(ber.ClassContext, ber.TypeConstructed, goldap.FilterOr, nil, goldap.FilterMap[goldap.FilterOr])
	p.AppendChild(f.packet)
	p.AppendChild(filterB.packet)
	return &FilterBuilder{packet: p}
}

// And ...
func (f *FilterBuilder) And(filterB *FilterBuilder) *FilterBuilder {
	if f.packet == nil {
		return filterB
	}
	if filterB.packet == nil {
		return f
	}
	p := ber.Encode(ber.ClassContext, ber.TypeConstructed, goldap.FilterAnd, nil, goldap.FilterMap[goldap.FilterAnd])
	p.AppendChild(f.packet)
	p.AppendChild(filterB.packet)
	return &FilterBuilder{packet: p}
}

// String ...
func (f *FilterBuilder) String() (string, error) {
	if f.packet == nil {
		return "", nil
	}
	return goldap.DecompileFilter(f.packet)
}

// NewFilterBuilder parse FilterBuilder from string
func NewFilterBuilder(filter string) (*FilterBuilder, error) {
	f := normalizeFilter(filter)
	if len(strings.TrimSpace(f)) == 0 {
		return &FilterBuilder{}, nil
	}
	p, err := goldap.CompileFilter(f)
	if err != nil {
		return &FilterBuilder{}, ErrInvalidFilter
	}
	return &FilterBuilder{packet: p}, nil
}

// normalizeFilter - add '(' and ')' in ldap filter if it doesn't exist
func normalizeFilter(filter string) string {
	norFilter := strings.TrimSpace(filter)
	if len(norFilter) == 0 {
		return norFilter
	}
	if strings.HasPrefix(norFilter, "(") && strings.HasSuffix(norFilter, ")") {
		return norFilter
	}
	return "(" + norFilter + ")"
}
