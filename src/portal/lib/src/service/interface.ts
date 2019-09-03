import { Project } from "../project-policy-config/project";
import { Observable } from 'rxjs';
import { ClrModal } from '@clr/angular';
import { HttpHeaders, HttpParams } from '@angular/common/http';

/**
 * The base interface contains the general properties
 *
 **
 * interface Base
 */
export interface Base {
  id?: string | number;
  name?: string;
  creation_time?: Date;
  update_time?: Date;
}

/**
 * Interface for Repository Info
 *
 **
 * interface Repository
 * extends {Base}
 */
export interface RepositoryItem extends Base {
  [key: string]: any | any[];
  name: string;
  tags_count: number;
  owner_id?: number;
  project_id?: number;
  description?: string;
  star_count?: number;
  pull_count?: number;
}

/**
 * Interface for repository
 *
 **
 * interface Repository
 */
export interface Repository {
  metadata?: Metadata;
  data: RepositoryItem[];
}

/**
 * Interface for the tag of repository
 *
 **
 * interface Tag
 * extends {Base}
 */

export interface Tag extends Base {
  digest: string;
  name: string;
  size: string;
  architecture: string;
  os: string;
  'os.version': string;
  docker_version: string;
  author: string;
  created: Date;
  signature?: string;
  scan_overview?: VulnerabilitySummary;
  labels: Label[];
  push_time?: string;
  pull_time?: string;
}

/**
 * Interface for registry endpoints.
 *
 **
 * interface Endpoint
 * extends {Base}
 */
export interface Endpoint extends Base {
  credential: {
    access_key?: string,
    access_secret?: string,
    type: string;
  };
  description: string;
  insecure: boolean;
  name: string;
  type: string;
  url: string;
}

export interface PingEndpoint extends Base {
  access_key?: string;
  access_secret?: string;
  description: string;
  insecure: boolean;
  name: string;
  type: string;
  url: string;
}

export interface Filter {
  type: string;
  style: string;
  values?: string[];
}

/**
 * Interface for replication rule.
 *
 **
 * interface ReplicationRule
 * interface Filter
 * interface Trigger
 */
export interface ReplicationRule extends Base {
  [key: string]: any;
  id?: number;
  name: string;
  description: string;
  trigger: Trigger;
  filters: Filter[];
  deletion?: boolean;
  src_registry?: any;
  dest_registry?: any;
  src_namespaces: string[];
  dest_namespace?: string;
  enabled: boolean;
  override: boolean;
}

export class Filter {
  type: string;
  value?: any;
  constructor(type: string) {
    this.type = type;
  }
}

export class Trigger {
  type: string;
  trigger_settings:
    | any
    | {
      [key: string]: any | any[];
    };
  constructor(type: string, param: any | { [key: string]: any | any[] }) {
    this.type = type;
    this.trigger_settings = param;
  }
}

/**
 * Interface for replication job.
 *
 **
 * interface ReplicationJob
 */
export interface ReplicationJob {
  metadata?: Metadata;
  data: ReplicationJobItem[];
}

/**
 * Interface for replication job item.
 *
 **
 * interface ReplicationJob
 */
export interface ReplicationJobItem extends Base {
  [key: string]: any | any[];
  id: number;
  status: string;
  policy_id: number;
  trigger: string;
  total: number;
  failed: number;
  succeed: number;
  in_progress: number;
  stopped: number;
}

/**
 * Interface for replication tasks item.
 *
 **
 * interface ReplicationTasks
 */
export interface ReplicationTasks extends Base {
  [key: string]: any | any[];
  operation: string;
  id: number;
  execution_id: number;
  resource_type: string;
  src_resource: string;
  dst_resource: string;
  job_id: number;
  status: string;
}
/**
 * Interface for storing metadata of response.
 *
 **
 * interface Metadata
 */
export interface Metadata {
  xTotalCount: number;
}

/**
 * Interface for access log.
 *
 **
 * interface AccessLog
 */
export interface AccessLog {
  metadata?: Metadata;
  data: AccessLogItem[];
}

/**
 * The access log data.
 *
 **
 * interface AccessLogItem
 */
export interface AccessLogItem {
  [key: string]: any | any[];
  log_id: number;
  project_id: number;
  repo_name: string;
  repo_tag: string;
  operation: string;
  op_time: string | Date;
  user_id: number;
  username: string;
  keywords?: string; // NOT used now
  guid?: string; // NOT used now
}

/**
 * Global system info.
 *
 **
 * interface SystemInfo
 *
 */
export interface SystemInfo {
  with_clair?: boolean;
  with_notary?: boolean;
  with_admiral?: boolean;
  with_chartmuseum?: boolean;
  admiral_endpoint?: string;
  auth_mode?: string;
  registry_url?: string;
  project_creation_restriction?: string;
  self_registration?: boolean;
  has_ca_root?: boolean;
  harbor_version?: string;
  clair_vulnerability_status?: ClairDBStatus;
  next_scan_all?: number;
  external_url?: string;
}

/**
 * Clair database status info.
 *
 **
 * interface ClairDetail
 */
export interface ClairDetail {
  namespace: string;
  last_update: number;
}

export interface ClairDBStatus {
  overall_last_update: number;
  details: ClairDetail[];
}

export enum VulnerabilitySeverity {
  _SEVERITY,
  NONE,
  UNKNOWN,
  LOW,
  MEDIUM,
  HIGH
}

export interface VulnerabilityBase {
  id: string;
  severity: VulnerabilitySeverity;
  package: string;
  version: string;
}

export interface VulnerabilityItem extends VulnerabilityBase {
  link: string;
  fixedVersion: string;
  layer?: string;
  description: string;
}

export interface VulnerabilitySummary {
  image_digest?: string;
  scan_status: string;
  job_id?: number;
  severity: VulnerabilitySeverity;
  components: VulnerabilityComponents;
  update_time: Date; // Use as complete timestamp
}

export interface VulnerabilityComponents {
  total: number;
  summary: VulnerabilitySeverityMetrics[];
}

export interface VulnerabilitySeverityMetrics {
  severity: VulnerabilitySeverity;
  count: number;
}

export interface TagClickEvent {
  project_id: string | number;
  repository_name: string;
  tag_name: string;
}

export interface Label {
  [key: string]: any | any[];
  name: string;
  description: string;
  color: string;
  scope: string;
  project_id: number;
}

export interface Quota {
  id: number;
  ref: {
    name: string;
    owner_name: string;
    id: number;
  } | null;
  creation_time: string;
  update_time: string;
  hard: {
    count: number;
    storage: number;
  };
  used: {
    count: number;
    storage: number;
  };
}
export interface QuotaHard {
  hard: QuotaCountStorage;
}
export interface QuotaCountStorage {
  count: number;
  storage: number;
}

export interface CardItemEvent {
  event_type: string;
  item: any;
  additional_info?: any;
}

export interface ScrollPosition {
  sH: number;
  sT: number;
  cH: number;
}

/**
 * The manifest of image.
 *
 **
 * interface Manifest
 */
export interface Manifest {
  manifset: Object;
  config: string;
}

export interface RetagRequest {
  targetProject: string;
  targetRepo: string;
  targetTag: string;
  srcImage: string;
  override: boolean;
}

export interface ClrDatagridComparatorInterface<T> {
  compare(a: T, b: T): number;
}

export interface ClrDatagridStateInterface {
  page?: { from?: number; to?: number; size?: number };
  sort?: { by: string | ClrDatagridComparatorInterface<any>; reverse: boolean };
  filters?: ({ property: string; value: string } | ClrDatagridFilterInterface<any>)[];
}

export interface ClrDatagridFilterInterface<T> {
  isActive(): boolean;

  accepts(item: T): boolean;

  changes: Observable<any>;
}

/** @deprecated since 0.11 */
export interface Comparator<T> extends ClrDatagridComparatorInterface<T> { }
/** @deprecated since 0.11 */
export interface ClrFilter<T> extends ClrDatagridFilterInterface<T> { }
/** @deprecated since 0.11 */
export interface State extends ClrDatagridStateInterface { }
export interface Modal extends ClrModal { }
export const Modal = ClrModal;

/**
 * The access user privilege from serve.
 *
 **
 * interface UserPrivilegeServe
 */
export interface UserPrivilegeServeItem {
  [key: string]: any | any[];
  resource: string;
  action: string;
}

export class OriginCron {
  type: string;
  cron: string;
}

export interface HttpOptionInterface {
  headers?: HttpHeaders | {
    [header: string]: string | string[];
  };
  observe?: 'body';
  params?: HttpParams | {
    [param: string]: string | string[];
  };
  reportProgress?: boolean;
  responseType: 'json';
  withCredentials?: boolean;
}

export interface HttpOptionTextInterface {
  headers?: HttpHeaders | {
    [header: string]: string | string[];
  };
  observe?: 'body';
  params?: HttpParams | {
    [param: string]: string | string[];
  };
  reportProgress?: boolean;
  responseType: 'text';
  withCredentials?: boolean;
}


export interface ProjectRootInterface {
  NAME: string;
  VALUE: number;
  LABEL: string;
}
export interface SystemCVEWhitelist {
  id: number;
  project_id: number;
  expires_at: number;
  items: Array<{ "cve_id": string; }>;
}
export interface QuotaHardInterface {
  count_per_project: number;
  storage_per_project: number;
}

export interface QuotaUnitInterface {
  UNIT: string;
}
export interface QuotaHardLimitInterface {
  countLimit: number;
  storageLimit: number;
  storageUnit: string;
  id?: string;
  countUsed?: string;
  storageUsed?: string;
}
export interface EditQuotaQuotaInterface {
  editQuota: string;
  setQuota: string;
  countQuota: string;
  storageQuota: string;
  quotaHardLimitValue: QuotaHardLimitInterface | any;
  isSystemDefaultQuota: boolean;
}
