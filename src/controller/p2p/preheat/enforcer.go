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

package preheat

import (
	"context"
	"fmt"
	"strings"

	tk "github.com/docker/distribution/registry/auth/token"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/controller/scan"
	"github.com/goharbor/harbor/src/controller/tag"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/service/token"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/lib/selector"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat"
	pol "github.com/goharbor/harbor/src/pkg/p2p/preheat/models/policy"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/models/provider"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/policy"
	pr "github.com/goharbor/harbor/src/pkg/p2p/preheat/provider"
	"github.com/goharbor/harbor/src/pkg/scan/report"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/goharbor/harbor/src/pkg/scan/vuln"
	"github.com/goharbor/harbor/src/pkg/task"
)

const (
	defaultSeverityCode     = 99
	extraAttrTotal          = "totalCount"
	extraAttrTrigger        = "trigger"
	extraAttrTriggerSetting = "triggerSetting"
	extraAttrArtifact       = "artifact"
	extraAttrDigest         = "digest"
	extraAttrKind           = "kind"
	resourceActionType      = "repository"
	resourcePullAction      = "pull"
	manifestAPIPattern      = "%s/v2/%s/manifests/%s"
	accessCredHeaderKey     = "Authorization"
)

// Enforcer defines preheat policy enforcement operations.
type Enforcer interface {
	// Enforce preheating action by the given policy.
	// For manual and scheduled preheating scenarios.
	//
	// Arguments:
	//   ctx context.Context : system context
	//   policyID int64 : ID of the being enforced policy
	//
	// Returns:
	//   - ID of the execution
	//   - non-nil error if any error occurred during the enforcement
	EnforcePolicy(ctx context.Context, policyID int64) (int64, error)

	// Enforce preheating action by the given artifact.
	// For event-based cases.
	// Using the given artifact to located the matched preheat policy and bound this action
	// with the located preheat policy.
	//
	// Arguments:
	//   ctx context.Context : system context
	//   art *artifact.Artifact: Artifact contained in the occurred events.
	//
	// Returns:
	//   - IDs of the executions
	//   - non-nil error if any error occurred during the enforcement
	//
	// Notes:
	//  The current design is artifact central mode (identified by digest). All the tags of
	//  the artifact are kept together. However, the preheating action is based on the specified
	//  tag and we need to split the all-tags-in-one artifact to one-tag artifacts here.
	PreheatArtifact(ctx context.Context, art *artifact.Artifact) ([]int64, error)
}

// extURLGetter is a func template to get the external access endpoint
// The purpose of defining such a func template is decoupling code
type extURLGetter func(c *selector.Candidate) (string, error)

// accessCredMaker is a func template to generate the required credential header value
// The purpose of defining such a func template is decoupling code
type accessCredMaker func(c *selector.Candidate) (string, error)

// defaultEnforcer is default implementation of Enforcer.
type defaultEnforcer struct {
	// for policy management
	policyMgr policy.Manager
	// for talking to job service to launch tasks
	executionMgr task.ExecutionManager
	taskMgr      task.Manager
	// for retrieving the artifact candidates (including tags and labels)
	artCtl artifact.Controller
	// for getting vulnerability severity of the specified artifact
	scanCtl scan.Controller
	// for getting project related info
	proCtl project.Controller
	// TODO: Need preheat provider manager
	//
	// for getting the access endpoint of registry V2
	fullURLGetter extURLGetter
	// for creating the access credential
	credMaker accessCredMaker
}

// NewEnforcer creat a new enforcer
func NewEnforcer() Enforcer {
	return &defaultEnforcer{
		policyMgr:    policy.New(),
		executionMgr: task.NewExecutionManager(),
		taskMgr:      task.NewManager(),
		artCtl:       artifact.NewController(),
		scanCtl:      scan.DefaultController,
		proCtl:       project.NewController(),
		fullURLGetter: func(c *selector.Candidate) (s string, e error) {
			edp, err := config.ExtEndpoint()
			if err != nil {
				return "", err
			}

			r := fmt.Sprintf("%s/%s", c.Namespace, c.Repository)
			return fmt.Sprintf(manifestAPIPattern, edp, r, c.Tags[0]), nil
		},
		credMaker: func(c *selector.Candidate) (s string, e error) {
			r := fmt.Sprintf("%s/%s", c.Namespace, c.Repository)

			ac := []*tk.ResourceActions{
				{
					Type: resourceActionType,
					Name: r,
					// Only pull action is enough
					Actions: []string{resourcePullAction},
				},
			}
			t, err := token.MakeToken("distributor", token.Registry, ac)
			if err != nil {
				return "", err
			}

			return fmt.Sprintf("Bearer %s", t.Token), nil
		},
	}
}

// EnforcePolicy enforces preheating action by the given policy
func (de *defaultEnforcer) EnforcePolicy(ctx context.Context, policyID int64) (int64, error) {
	// Get the the given policy data
	pl, err := de.policyMgr.Get(ctx, policyID)
	if err != nil {
		return -1, enforceError(err)
	}

	// Check if policy is enabled
	if !pl.Enabled {
		return -1, enforceError(errors.Errorf("policy %d:%s is not enabled", pl.ID, pl.Name))
	}

	// Retrieve the initial candidates
	candidates, err := de.getCandidates(ctx, pl)
	if err != nil {
		return -1, enforceError(err)
	}

	// Do filters
	filtered, err := policy.NewFilter().
		BuildFrom(pl).
		Filter(candidates)
	if err != nil {
		return -1, enforceError(err)
	}

	// Launch execution
	eid, err := de.launchExecutions(ctx, filtered, pl)
	if err != nil {
		// NOTES: Please pay attention here, even the non-nil error returned, it does not mean
		// the relevant execution is not available. The execution ID should also be checked(>0)
		// at any time.
		return eid, enforceError(err)
	}

	return eid, nil
}

// PreheatArtifact enforces preheating action by the given artifact.
func (de *defaultEnforcer) PreheatArtifact(ctx context.Context, art *artifact.Artifact) ([]int64, error) {
	if art == nil {
		return nil, errors.New("nil artifact")
	}

	// Get project info
	p, err := de.proCtl.Get(ctx, art.ProjectID, project.CVEAllowlist(true))
	if err != nil {
		return nil, enforceErrorExt(err, art)
	}

	// Convert to candidates
	candidates, err := de.toCandidates(ctx, p, []*artifact.Artifact{art})
	if err != nil {
		return nil, enforceErrorExt(err, art)
	}

	// Find all the policies that match the given artifact
	l, err := de.policyMgr.ListPoliciesByProject(ctx, art.ProjectID, nil)
	if err != nil {
		return nil, enforceErrorExt(err, art)
	}

	matched := 0
	ids := make([]int64, 0)
	for _, pl := range l {
		// Skip disabled policies
		if !pl.Enabled {
			continue
		}

		// Only look for the event-based policies
		if pl.Trigger == nil ||
			pl.Trigger.Type != pol.TriggerTypeEventBased {
			// Skip
			continue
		}

		filtered, err := policy.NewFilter().BuildFrom(pl).Filter(candidates)
		if err != nil {
			// Log error and continue
			logger.Errorf("Failed to do filter for policy %d:%s with error: %s", pl.ID, pl.Name, err.Error())
			continue
		}

		matched++

		if len(filtered) > 0 {
			// Matched
			eid, err := de.launchExecutions(ctx, filtered, pl)
			if err != nil {
				// Log error and continue
				logger.Errorf("Failed to launch execution for policy %d:%s with error: %s", pl.ID, pl.Name, err.Error())
			} else {
				// Success and then append the execution id to list
				ids = append(ids, eid)
			}
		}
	}

	if matched != len(ids) {
		// Some policy enforcement are failed
		// Treat it as an error case
		return ids, enforceErrorExt(errors.Errorf("%d policies matched but only %d successfully enforced", matched, len(ids)), art)
	}

	return ids, nil
}

// getCandidates get the initial candidates by evaluating the policy
func (de *defaultEnforcer) getCandidates(ctx context.Context, ps *pol.Schema) ([]*selector.Candidate, error) {
	// Get project info
	p, err := de.proCtl.Get(ctx, ps.ProjectID, project.CVEAllowlist(true))
	if err != nil {
		return nil, err
	}

	// Get the initial candidates
	// Here we have a hidden filter, the artifact type filter.
	// Only get the image type at this moment.
	arts, err := de.artCtl.List(ctx, &q.Query{
		Keywords: map[string]interface{}{
			"project_id": ps.ProjectID,
			"type":       pr.SupportedType,
		},
	}, &artifact.Option{
		WithLabel: true,
		WithTag:   true,
		TagOption: &tag.Option{
			WithSignature: true,
		},
	})
	if err != nil {
		return nil, err
	}

	return de.toCandidates(ctx, p, arts)
}

// launchExecutions create execution record and launch tasks to preheat the filtered artifacts.
func (de *defaultEnforcer) launchExecutions(ctx context.Context, candidates []*selector.Candidate, pl *pol.Schema) (int64, error) {
	// Create execution first anyway
	attrs := map[string]interface{}{
		extraAttrTotal:          len(candidates),
		extraAttrTrigger:        pl.Trigger.Type,
		extraAttrTriggerSetting: pl.Trigger.Settings.Cron,
	}
	if pl.Trigger.Type != pol.TriggerTypeScheduled {
		attrs[extraAttrTriggerSetting] = "-"
	}

	eid, err := de.executionMgr.Create(ctx, job.P2PPreheat, pl.ID, pl.Trigger.Type, attrs)
	if err != nil {
		return -1, err
	}

	// Handle empty candidate list case
	if len(candidates) == 0 {
		// Return earlier
		if err := de.executionMgr.MarkDone(ctx, eid, "no artifacts to preheat"); err != nil {
			return eid, err
		}

		return eid, nil
	}

	// TODO: Get provider instance by the provider ID
	// Placeholder
	ins := &provider.Instance{}
	insData, err := ins.ToJSON()
	if err != nil {
		// In case
		if er := de.executionMgr.MarkError(ctx, eid, err.Error()); er != nil {
			return eid, errors.Wrap(er, err.Error())
		}

		return eid, nil
	}

	// Start tasks
	count := 0
	for _, c := range candidates {
		if _, err = de.startTask(ctx, eid, c, insData); err != nil {
			// Just log the error and skip
			logger.Errorf("start task error for preheating image: %s/%s:%s@%s", c.Namespace, c.Repository, c.Tags[0], c.Digest)
			continue
		}

		count++
	}

	if count != len(candidates) {
		// Obviously, failed to start some tasks
		// Return as an error but the execution can still be queried.
		return eid, errors.Errorf("some errors occurred when enforcing policy '%s(%d)' but execution '%d' is still available", pl.Name, pl.ID, eid)
	}

	return eid, nil
}

// startTask starts the preheat task(job) for the given candidate
func (de *defaultEnforcer) startTask(ctx context.Context, executionID int64, candidate *selector.Candidate, instance string) (int64, error) {
	u, err := de.fullURLGetter(candidate)
	if err != nil {
		return -1, err
	}

	cred, err := de.credMaker(candidate)
	if err != nil {
		return -1, err
	}

	pi := &pr.PreheatImage{
		Type: pr.SupportedType,
		URL:  u,
		Headers: map[string]interface{}{
			accessCredHeaderKey: cred,
		},
		ImageName: fmt.Sprintf("%s/%s", candidate.Namespace, candidate.Repository),
		Tag:       candidate.Tags[0],
	}

	piData, err := pi.ToJSON()
	if err != nil {
		return -1, err
	}

	j := &task.Job{
		Name: job.P2PPreheat,
		Parameters: job.Parameters{
			preheat.PreheatParamProvider: instance,
			preheat.PreheatParamImage:    piData,
		},
		Metadata: &job.Metadata{
			JobKind:  job.KindGeneric,
			IsUnique: true,
		},
	}

	tid, err := de.taskMgr.Create(ctx, executionID, j, map[string]interface{}{
		extraAttrArtifact: fmt.Sprintf("%s:%s", pi.ImageName, pi.Tag),
		extraAttrDigest:   candidate.Digest,
		extraAttrKind:     pi.Type,
	})
	if err != nil {
		return -1, err
	}

	return tid, nil
}

// getVulnerabilitySev gets the severity code value for the given artifact with allowlist option set
func (de *defaultEnforcer) getVulnerabilitySev(ctx context.Context, p *models.Project, art *artifact.Artifact) (uint, error) {
	al := report.CVESet(p.CVEAllowlist.CVESet())
	r, err := de.scanCtl.GetSummary(ctx, art, []string{v1.MimeTypeNativeReport}, report.WithCVEAllowlist(&al))
	if err != nil {
		if errors.IsNotFoundErr(err) {
			// no vulnerability report
			return defaultSeverityCode, nil
		}

		return defaultSeverityCode, errors.Wrap(err, "get vulnerability severity")
	}

	// Severity is based on the native report format.
	// In case no supported report format, treat as same to the no report scenario
	sum, ok := r[v1.MimeTypeNativeReport]
	if !ok {
		return defaultSeverityCode, nil
	}

	sm, ok := sum.(*vuln.NativeReportSummary)
	if !ok {
		return defaultSeverityCode, errors.New("malformed native summary report")
	}

	return (uint)(sm.Severity.Code()), nil
}

// toCandidates converts the artifacts to filtering candidates
func (de *defaultEnforcer) toCandidates(ctx context.Context, p *models.Project, arts []*artifact.Artifact) ([]*selector.Candidate, error) {
	// Convert to filtering candidates first
	candidates := make([]*selector.Candidate, 0)

	for _, a := range arts {
		// Vulnerability severity is property of artifact
		sev, err := de.getVulnerabilitySev(ctx, p, a)
		if err != nil {
			return nil, err
		}

		// If artifact has more than one tag, then split them into separate candidate for easy filtering.
		for _, t := range a.Tags {
			candidates = append(candidates, &selector.Candidate{
				NamespaceID: p.ProjectID,
				Namespace:   p.Name,
				Repository:  pureRepository(p.Name, a.RepositoryName),
				Kind:        pr.SupportedType,
				Digest:      a.Digest,
				Tags:        []string{t.Name},
				Labels:      getLabels(a.Labels),
				Signatures: map[string]bool{
					t.Name: t.Signed,
				},
				VulnerabilitySeverity: sev,
			})
		}
	}

	return candidates, nil
}

// enforceError is a wrap error
func enforceError(e error) error {
	return errors.Wrap(e, "enforce policy error")
}

// enforceErrorExt is an enhanced wrap error
func enforceErrorExt(e error, art *artifact.Artifact) error {
	return errors.Wrap(e, fmt.Sprintf("enforce policy for given artifact error: %s@%s", art.RepositoryName, art.Digest))
}

// pureRepository removes project name from the repository
func pureRepository(ns, r string) string {
	return strings.TrimPrefix(r, fmt.Sprintf("%s/", ns))
}

// getLabels gets label texts from the label objects
func getLabels(labels []*models.Label) []string {
	lt := make([]string, 0)
	for _, l := range labels {
		lt = append(lt, l.Name)
	}

	return lt
}
