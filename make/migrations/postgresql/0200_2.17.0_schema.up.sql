/*
Add create_repo_if_not_exist to replication_policy.

Controls whether Harbor pre-creates the repository on the destination registry
before pushing during replication. NULL/true preserves current behavior
(Harbor creates the repo if it doesn't exist). false lets the destination
registry auto-create it on push instead, so registry-side automation such as
AWS ECR repository creation templates can apply.

See: https://github.com/goharbor/harbor/issues/22842
*/
ALTER TABLE replication_policy ADD COLUMN IF NOT EXISTS create_repo_if_not_exist boolean;
