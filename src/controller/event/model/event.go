package model

import "github.com/goharbor/harbor/src/pkg/retention/policy/rule"

// Replication describes replication infos
type Replication struct {
	HarborHostname     string               `json:"harbor_hostname,omitempty"`
	JobStatus          string               `json:"job_status,omitempty"`
	Description        string               `json:"description,omitempty"`
	ArtifactType       string               `json:"artifact_type,omitempty"`
	AuthenticationType string               `json:"authentication_type,omitempty"`
	OverrideMode       bool                 `json:"override_mode,omitempty"`
	TriggerType        string               `json:"trigger_type,omitempty"`
	PolicyCreator      string               `json:"policy_creator,omitempty"`
	ExecutionTimestamp int64                `json:"execution_timestamp,omitempty"`
	SrcResource        *ReplicationResource `json:"src_resource,omitempty"`
	DestResource       *ReplicationResource `json:"dest_resource,omitempty"`
	SuccessfulArtifact []*ArtifactInfo      `json:"successful_artifact,omitempty"`
	FailedArtifact     []*ArtifactInfo      `json:"failed_artifact,omitempty"`
}

// ArtifactInfo describe info of artifact
type ArtifactInfo struct {
	Type       string `json:"type"`
	Status     string `json:"status"`
	NameAndTag string `json:"name_tag"`
	FailReason string `json:"fail_reason,omitempty"`
}

// ReplicationResource describes replication resource info
type ReplicationResource struct {
	RegistryName string `json:"registry_name,omitempty"`
	RegistryType string `json:"registry_type"`
	Endpoint     string `json:"endpoint"`
	Provider     string `json:"provider,omitempty"`
	Namespace    string `json:"namespace,omitempty"`
}

// Retention describes tag retention infos
type Retention struct {
	Total             int              `json:"total"`
	Retained          int              `json:"retained"`
	HarborHostname    string           `json:"harbor_hostname,omitempty"`
	ProjectName       string           `json:"project_name,omitempty"`
	RetentionPolicyID int64            `json:"retention_policy_id,omitempty"`
	RetentionRules    []*RetentionRule `json:"retention_rule,omitempty"`
	Status            string           `json:"result,omitempty"`
	DeletedArtifact   []*ArtifactInfo  `json:"deleted_artifact,omitempty"`
}

// RetentionRule describes tag retention rule
type RetentionRule struct {
	// Template ID
	Template string `json:"template,omitempty"`
	// The parameters of this rule
	Parameters map[string]rule.Parameter `json:"params,omitempty"`
	// Selector attached to the rule for filtering tags
	TagSelectors []*rule.Selector `json:"tag_selectors,omitempty" `
	// Selector attached to the rule for filtering scope (e.g: repositories or namespaces)
	ScopeSelectors map[string][]*rule.Selector `json:"scope_selectors,omitempty"`
}
