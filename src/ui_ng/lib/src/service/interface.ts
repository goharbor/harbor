/**
 * The base interface contains the general properties
 * 
 * @export
 * @interface Base
 */
export interface Base {
    id?: string | number;
    name?: string;
    creation_time?: Date;
    update_time?: Date;
}

/**
 * Interface for tag history
 * 
 * @export
 * @interface TagCompatibility
 */
export interface TagCompatibility {
    v1Compatibility: string;
}

/**
 * Interface for tag manifest
 * 
 * @export
 * @interface TagManifest
 */
export interface TagManifest {
    schemaVersion: number;
    name: string;
    tag: string;
    architecture: string;
    history: TagCompatibility[];
}

/**
 * Interface for Repository
 * 
 * @export
 * @interface Repository
 * @extends {Base}
 */
export interface Repository extends Base {
    name: string;
    tags_count: number;
    owner_id?: number;
    project_id?: number;
    description?: string;
    start_count?: number;
    pull_count?: number;
}

/**
 * Interface for the tag of repository
 * 
 * @export
 * @interface Tag
 * @extends {Base}
 */
export interface Tag extends Base {
    tag: string;
    manifest: TagManifest;
    signed?: number; //May NOT exist
}

/**
 * Interface for registry endpoints.
 * 
 * @export
 * @interface Endpoint
 * @extends {Base}
 */
export interface Endpoint extends Base {
  endpoint: string;
  name: string;
  username: string;
  password: string;
  type: number;
}

/**
 * Interface for replication rule.
 * 
 * @export
 * @interface ReplicationRule
 */
export interface ReplicationRule extends Base {
    project_id: number;
    project_name: string;
    target_id: number;
    target_name: string;
    enabled: number;
    description?: string;
    cron_str?: string;
    start_time?: Date;
    error_job_count?: number;
    deleted: number;
}

/**
 * Interface for replication job.
 * 
 * @export
 * @interface ReplicationJob
 */
export interface ReplicationJob extends Base {
    status: string;
    repository: string;
    policy_id: number;
    operation: string;
    tags: string;
}

/**
 * Interface for access log.
 * 
 * @export
 * @interface AccessLog
 */
export interface AccessLog {
    log_id: number;
    project_id: number;
    repo_name: string;
    repo_tag: string;
    operation: string;
    op_time: string | Date;
    user_id: number;
    username: string;
    keywords?: string; //NOT used now
    guid?: string; //NOT used now
}