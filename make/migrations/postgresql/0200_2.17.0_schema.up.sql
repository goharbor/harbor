/*
Add adapter_options to replication_policy.

A generic JSON-encoded key/value bag for adapter-specific settings that don't
warrant a dedicated column of their own. Each registry adapter reads only the
keys it understands and ignores the rest.

Initial use case: the AWS ECR adapter reads "skip_repo_creation": "true" to
skip pre-creating the destination repository before pushing, so that ECR
repository creation templates -- which only apply when ECR creates the
repository itself, not when it's created via an explicit CreateRepository
API call -- can take effect.

See: https://github.com/goharbor/harbor/issues/22842
*/
ALTER TABLE replication_policy ADD COLUMN IF NOT EXISTS adapter_options text;
