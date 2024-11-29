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

package policy

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// PolicyTestSuite is a test suite for policy schema.
type PolicyTestSuite struct {
	suite.Suite

	schema *Schema
}

// TestPolicy is the entry method of running PolicyTestSuite.
func TestPolicy(t *testing.T) {
	suite.Run(t, &PolicyTestSuite{})
}

// SetupSuite prepares the env for PolicyTestSuite.
func (p *PolicyTestSuite) SetupSuite() {
	p.schema = &Schema{}
	p.schema.Trigger = &Trigger{}
}

// TearDownSuite clears the env for PolicyTestSuite.
func (p *PolicyTestSuite) TearDownSuite() {
	p.schema = nil
}

// TestValidatePreheatPolicy tests the ValidatePreheatPolicy method
func (p *PolicyTestSuite) TestValidatePreheatPolicy() {
	// manual trigger
	p.schema.Trigger.Type = TriggerTypeManual
	p.NoError(p.schema.ValidatePreheatPolicy())

	// event trigger
	p.schema.Trigger.Type = TriggerTypeEventBased
	p.NoError(p.schema.ValidatePreheatPolicy())

	// scheduled trigger
	p.schema.Trigger.Type = TriggerTypeScheduled
	// cron string is empty
	p.schema.Trigger.Settings.Cron = ""
	p.NoError(p.schema.ValidatePreheatPolicy())
	// the 1st field of cron string is not 0
	p.schema.Trigger.Settings.Cron = "1 0 0 1 1 *"
	p.Error(p.schema.ValidatePreheatPolicy())
	// valid cron string
	p.schema.Trigger.Settings.Cron = "0 0 0 1 1 *"
	p.NoError(p.schema.ValidatePreheatPolicy())

	// invalid preheat scope
	p.schema.Scope = "invalid scope"
	p.Error(p.schema.ValidatePreheatPolicy())
	// valid preheat scope
	p.schema.Scope = "single_peer"
	p.NoError(p.schema.ValidatePreheatPolicy())
}

// TestDecode tests decode.
func (p *PolicyTestSuite) TestDecode() {
	s := &Schema{
		ID:            100,
		Name:          "test-for-decode",
		Description:   "",
		ProjectID:     1,
		ProviderID:    1,
		Filters:       nil,
		FiltersStr:    "[{\"type\":\"repository\",\"value\":\"**\"},{\"type\":\"tag\",\"value\":\"**\"},{\"type\":\"label\",\"value\":\"test\"}]",
		Trigger:       nil,
		TriggerStr:    "{\"type\":\"event_based\",\"trigger_setting\":{\"cron\":\"\"}}",
		Enabled:       false,
		Scope:         "all_peers",
		ExtraAttrsStr: "{\"key\":\"value\"}",
	}
	p.NoError(s.Decode())
	p.Len(s.Filters, 3)
	p.NotNil(s.Trigger)

	p.Equal(ScopeTypeAllPeers, s.Scope)
	p.Equal(map[string]interface{}{"key": "value"}, s.ExtraAttrs)

	// invalid filter or trigger
	s.FiltersStr = ""
	s.TriggerStr = "invalid"
	p.Error(s.Decode())

	s.FiltersStr = "invalid"
	s.TriggerStr = ""
	p.Error(s.Decode())
}

// TestEncode tests encode.
func (p *PolicyTestSuite) TestEncode() {
	s := &Schema{
		ID:          101,
		Name:        "test-for-encode",
		Description: "",
		ProjectID:   2,
		ProviderID:  2,
		Filters: []*Filter{
			{
				Type:  FilterTypeRepository,
				Value: "**",
			},
			{
				Type:  FilterTypeTag,
				Value: "**",
			},
			{
				Type:  FilterTypeLabel,
				Value: "test",
			},
		},
		FiltersStr: "",
		Trigger: &Trigger{
			Type: "event_based",
		},
		TriggerStr: "",
		Enabled:    false,
		Scope:      "single_peer",
		ExtraAttrs: map[string]interface{}{
			"key": "value",
		},
	}
	p.NoError(s.Encode())
	p.Equal(`[{"type":"repository","value":"**"},{"type":"tag","value":"**"},{"type":"label","value":"test"}]`, s.FiltersStr)
	p.Equal(`{"type":"event_based","trigger_setting":{}}`, s.TriggerStr)
	p.Equal(ScopeTypeSinglePeer, s.Scope)
	p.Equal(`{"key":"value"}`, s.ExtraAttrsStr)
}
