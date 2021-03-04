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

package legacy

import "github.com/goharbor/harbor/src/pkg/scheduler"

// As one job implementation can only be registered with one name. Define the following three
// schedulers which are wrapper of pkg/scheduler.PeriodicJob for the legacy periodical jobs
// They can be removed after several releases

// ReplicationScheduler is the legacy scheduler for replication
type ReplicationScheduler struct {
	scheduler.PeriodicJob
}

// GarbageCollectionScheduler is the legacy scheduler for garbage collection
type GarbageCollectionScheduler struct {
	scheduler.PeriodicJob
}

// ScanAllScheduler is the legacy scheduler for scan all
type ScanAllScheduler struct {
	scheduler.PeriodicJob
}
