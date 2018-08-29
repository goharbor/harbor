// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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

package models

import (
	"testing"

	"github.com/astaxie/beego/validation"
	"github.com/goharbor/harbor/src/replication"
	"github.com/stretchr/testify/assert"
)

func TestValidOfTrigger(t *testing.T) {
	cases := map[*Trigger]bool{
		{}: true,
		{
			Kind: "invalid_kind",
		}: true,
		{
			Kind: replication.TriggerKindImmediate,
		}: false,
		{
			Kind: replication.TriggerKindSchedule,
		}: true,
	}

	for filter, hasError := range cases {
		v := &validation.Validation{}
		filter.Valid(v)
		assert.Equal(t, hasError, v.HasErrors())
	}
}

func TestValidOfScheduleParam(t *testing.T) {
	cases := map[*ScheduleParam]bool{
		{}: true,
		{
			Type: "invalid_type",
		}: true,
		{
			Type:    replication.TriggerScheduleDaily,
			Offtime: 3600*24 + 1,
		}: true,
		{
			Type:    replication.TriggerScheduleDaily,
			Offtime: 3600 * 2,
		}: false,
		{
			Type:    replication.TriggerScheduleWeekly,
			Weekday: 0,
			Offtime: 3600 * 2,
		}: true,
		{
			Type:    replication.TriggerScheduleWeekly,
			Weekday: 7,
			Offtime: 3600 * 2,
		}: false,
	}

	for param, hasError := range cases {
		v := &validation.Validation{}
		param.Valid(v)
		assert.Equal(t, hasError, v.HasErrors())
	}
}
