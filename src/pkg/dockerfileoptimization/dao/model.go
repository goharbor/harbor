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

package dao

import (
	"time"

	"github.com/beego/beego/v2/client/orm"
)

func init() {
	orm.RegisterModel(&DockerfileOptimization{})
}

// Status values of an optimization record. Rows are created as Pending when the
// optimize job is launched and move to a terminal state when the adapter reports back.
const (
	StatusPending = "Pending"
	StatusRunning = "Running"
	StatusSuccess = "Success"
	StatusError   = "Error"
)

// DockerfileOptimization holds the persisted result of a Dockerfile extraction + LLM optimization.
type DockerfileOptimization struct {
	ID                        int64  `orm:"pk;auto;column(id)"`
	RepositoryName            string `orm:"column(repository_name)"`
	ArtifactDigest            string `orm:"column(artifact_digest)"`
	Dockerfile                string `orm:"column(dockerfile)"`
	OptimizedDockerfile       string `orm:"column(optimized_dockerfile)"`
	AttestationManifestDigest string `orm:"column(attestation_manifest_digest)"`
	StatementDigest           string `orm:"column(statement_digest)"`
	// Generated is true when Dockerfile was reconstructed from the image's
	// config history because no BuildKit provenance attestation was found,
	// rather than extracted verbatim from one.
	Generated        bool      `orm:"column(generated)"`
	Status           string    `orm:"column(status)"`
	Error            string    `orm:"column(error)"`
	RegistrationUUID string    `orm:"column(registration_uuid)"`
	ExecutionID      int64     `orm:"column(execution_id)"`
	CreatedAt        time.Time `orm:"column(created_at);auto_now_add"`
	UpdateTime       time.Time `orm:"column(update_time);auto_now"`
}

// TableName returns the database table name.
func (d *DockerfileOptimization) TableName() string {
	return "dockerfile_optimization"
}
