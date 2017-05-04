/**
 * The base interface contains the general properties
 * 
 * @export
 * @interface Base
 */
export interface Base {
    id?: string;
    name?: string;
    creation_time?: Date;
    update_time?: Date;
}

/**
 * Interface for Repository
 * 
 * @export
 * @interface Repository
 * @extends {Base}
 */
export interface Repository extends Base { }

/**
 * Interface for the tag of repository
 * 
 * @export
 * @interface Tag
 * @extends {Base}
 */
export interface Tag extends Base { }

/**
 * Interface for registry endpoints.
 * 
 * @export
 * @interface Endpoint
 * @extends {Base}
 */
export interface Endpoint extends Base { }

/**
 * Interface for replication rule.
 * 
 * @export
 * @interface ReplicationRule
 */
export interface ReplicationRule { }

/**
 * Interface for replication job.
 * 
 * @export
 * @interface ReplicationJob
 */
export interface ReplicationJob { }

/**
 * Interface for access log.
 * 
 * @export
 * @interface AccessLog
 */
export interface AccessLog {
    log_id: number,
    project_id: number,
    repo_name: string,
    repo_tag: string,
    operation: string,
    op_time: string | Date,
    user_id: number,
    username: string,
    keywords?: string, //NOT used now
    guid?: string //NOT used now
}