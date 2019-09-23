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

package vuln

const (
	// Unknown - either a security problem that has not been assigned to a priority yet or
	// a priority that the scanner did not recognize.
	Unknown Severity = "Unknown"
	// Low - a security problem, but is hard to exploit due to environment, requires a
	// user-assisted attack, a small install base, or does very little damage.
	Low Severity = "Low"
	// Negligible - technically a security problem, but is only theoretical in nature, requires
	// a very special situation, has almost no install base, or does no real damage.
	Negligible Severity = "Negligible"
	// Medium - a real security problem, and is exploitable for many people. Includes network
	// daemon denial of service attacks, cross-site scripting, and gaining user privileges.
	Medium Severity = "Medium"
	// High - a real problem, exploitable for many people in a default installation. Includes
	// serious remote denial of service, local root privilege escalations, or data loss.
	High Severity = "High"
	// Critical - a world-burning problem, exploitable for nearly all people in a default installation.
	// Includes remote root privilege escalations, or massive data loss.
	Critical Severity = "Critical"
)

// Severity is a standard scale for measuring the severity of a vulnerability.
type Severity string
