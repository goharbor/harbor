// Copyright 2015 clair authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package types defines useful types that are used in database models.
package types

import (
	"database/sql/driver"
	"errors"
	"fmt"
)

// Priority defines a vulnerability priority
type Priority string

const (
	// Unknown is either a security problem that has not been
	// assigned to a priority yet or a priority that our system
	// did not recognize
	Unknown Priority = "Unknown"
	// Negligible is technically a security problem, but is
	// only theoretical in nature, requires a very special
	// situation, has almost no install base, or does no real
	// damage. These tend not to get backport from upstreams,
	// and will likely not be included in security updates unless
	// there is an easy fix and some other issue causes an update.
	Negligible Priority = "Negligible"
	// Low is a security problem, but is hard to
	// exploit due to environment, requires a user-assisted
	// attack, a small install base, or does very little damage.
	// These tend to be included in security updates only when
	// higher priority issues require an update, or if many
	// low priority issues have built up.
	Low Priority = "Low"
	// Medium is a real security problem, and is exploitable
	// for many people. Includes network daemon denial of service
	// attacks, cross-site scripting, and gaining user privileges.
	// Updates should be made soon for this priority of issue.
	Medium Priority = "Medium"
	// High is a real problem, exploitable for many people in a default
	// installation. Includes serious remote denial of services,
	// local root privilege escalations, or data loss.
	High Priority = "High"
	// Critical is a world-burning problem, exploitable for nearly all people
	// in a default installation of Linux. Includes remote root
	// privilege escalations, or massive data loss.
	Critical Priority = "Critical"
	// Defcon1 is a Critical problem which has been manually highlighted by
	// the team. It requires an immediate attention.
	Defcon1 Priority = "Defcon1"
)

// Priorities lists all known priorities, ordered from lower to higher
var Priorities = []Priority{Unknown, Negligible, Low, Medium, High, Critical, Defcon1}

// IsValid determines if the priority is a valid one
func (p Priority) IsValid() bool {
	for _, pp := range Priorities {
		if p == pp {
			return true
		}
	}

	return false
}

// Compare compares two priorities
func (p Priority) Compare(p2 Priority) int {
	var i1, i2 int

	for i1 = 0; i1 < len(Priorities); i1 = i1 + 1 {
		if p == Priorities[i1] {
			break
		}
	}
	for i2 = 0; i2 < len(Priorities); i2 = i2 + 1 {
		if p2 == Priorities[i2] {
			break
		}
	}

	return i1 - i2
}

func (p *Priority) Scan(value interface{}) error {
	val, ok := value.([]byte)
	if !ok {
		return errors.New("could not scan a Priority from a non-string input")
	}
	*p = Priority(string(val))
	if !p.IsValid() {
		return fmt.Errorf("could not scan an invalid Priority (%v)", p)
	}
	return nil
}

func (p *Priority) Value() (driver.Value, error) {
	return string(*p), nil
}
