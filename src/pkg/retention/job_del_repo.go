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

package retention

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/pkg/retention/dep"
	"github.com/goharbor/harbor/src/pkg/retention/res"
	"github.com/olekukonko/tablewriter"
)

// DelRepoJob tries to delete the whole given repository
type DelRepoJob struct{}

// MaxFails of the job
func (drj *DelRepoJob) MaxFails() uint {
	return 3
}

// ShouldRetry indicates job can be retried if failed
func (drj *DelRepoJob) ShouldRetry() bool {
	return true
}

// Validate the parameters
func (drj *DelRepoJob) Validate(params job.Parameters) (err error) {
	if _, err = getParamRepo(params); err == nil {
		_, err = getParamDryRun(params)
	}

	return
}

// Run the job
func (drj *DelRepoJob) Run(ctx job.Context, params job.Parameters) error {
	// logger for logging
	myLogger := ctx.GetLogger()

	// Parameters have been validated, ignore error checking
	repo, _ := getParamRepo(params)
	isDryRun, _ := getParamDryRun(params)

	// Log stage: start
	repoPath := fmt.Sprintf("%s/%s", repo.Namespace, repo.Name)
	myLogger.Infof("Run retention process.\n Repository: %s \n Dry Run: %v", repoPath, isDryRun)

	// For printing retention log
	allArtifacts, err := dep.DefaultClient.GetCandidates(repo)
	if err != nil {
		return err
	}

	// Stop check point:
	if isStopped(ctx) {
		logStop(myLogger)
		return nil
	}

	// Delete the repository
	if !isDryRun {
		if err := dep.DefaultClient.DeleteRepository(repo); err != nil {
			return err
		}
	}

	// Log deletions
	logDeletions(myLogger, allArtifacts)

	return nil
}

func logDeletions(logger logger.Interface, all []*res.Candidate) {
	var buf bytes.Buffer

	data := make([][]string, len(all))
	for _, c := range all {
		row := []string{
			arn(c),
			c.Kind,
			strings.Join(c.Labels, ","),
			t(c.PushedTime),
			t(c.PulledTime),
			t(c.CreationTime),
			actionMarkDeletion,
		}
		data = append(data, row)
	}

	table := tablewriter.NewWriter(&buf)
	table.SetHeader([]string{"Artifact", "Kind", "labels", "PushedTime", "PulledTime", "CreatedTime", "Retention"})
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.AppendBulk(data)
	table.Render()

	logger.Infof("\n%s", buf.String())
}
