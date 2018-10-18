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

package utils

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"

	"github.com/gocraft/work"
)

// Functions defined here are mainly from dep lib "github.com/gocraft/work".
// Only for compatible

// MakeIdentifier creates uuid for job.
func MakeIdentifier() string {
	b := make([]byte, 12)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%x", b)
}

// MakeUniquePeriodicID creates id for the periodic job.
func MakeUniquePeriodicID(name, spec string, epoch int64) string {
	return fmt.Sprintf("periodic:job:%s:%s:%d", name, spec, epoch)
}

// RedisNamespacePrefix ... Same with 'KeyNamespacePrefix', only for compatibility.
func RedisNamespacePrefix(namespace string) string {
	return KeyNamespacePrefix(namespace)
}

// RedisKeyScheduled returns key of scheduled job.
func RedisKeyScheduled(namespace string) string {
	return RedisNamespacePrefix(namespace) + "scheduled"
}

// RedisKeyLastPeriodicEnqueue returns key of timestamp if last periodic enqueue.
func RedisKeyLastPeriodicEnqueue(namespace string) string {
	return RedisNamespacePrefix(namespace) + "last_periodic_enqueue"
}

// RedisKeyDead returns key of the dead jobs.
func RedisKeyDead(namespace string) string {
	return RedisNamespacePrefix(namespace) + "dead"
}

// SerializeJob encodes work.Job to json data.
func SerializeJob(job *work.Job) ([]byte, error) {
	return json.Marshal(job)
}

// DeSerializeJob decodes bytes to ptr of work.Job.
func DeSerializeJob(jobBytes []byte) (*work.Job, error) {
	var j work.Job
	err := json.Unmarshal(jobBytes, &j)

	return &j, err
}
