import { Project } from "../project-policy-config/project";
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
  docker_version: string;
  author: string;
  created: Date;
  signature?: string;
  scan_overview?: VulnerabilitySummary;
  labels: Label[];
}

/**
 * Interface for registry endpoints.
 *
 **
 * interface Endpoint
 * extends {Base}
 */
export interface Endpoint extends Base {
  endpoint: string;
  name: string;
  username?: string;
  password?: string;
  insecure: boolean;
  type: number;
  [key: string]: any;
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
  projects: Project[];
  targets: Endpoint[];
  trigger: Trigger;
  filters: Filter[];
  replicate_existing_image_now?: boolean;
  replicate_deletion?: boolean;
}

export class Filter {
  kind: string;
  pattern: string;
  constructor(kind: string, pattern: string) {
    this.kind = kind;
    this.pattern = pattern;
  }
}

export class Trigger {
  kind: string;
  schedule_param:
    | any
    | {
        [key: string]: any | any[];
      };
  constructor(kind: string, param: any | { [key: string]: any | any[] }) {
    this.kind = kind;
    this.schedule_param = param;
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
  status: string;
  repository: string;
  policy_id: number;
  operation: string;
  tags: string;
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
export interface HelmChartSearchResultItem {
  Name: string;
  Score: number;
  Chart: HelmChartVersion;
}
export interface HelmChartItem {
  name: string;
  total_versions: number;
  latest_version: string;
  created: string;
  icon: string;
  home: string;
  status?: string;
  pulls?: number;
  maintainer?: string;
  deprecated?: boolean;
}

export interface HelmChartVersion {
  name: string;
  home: string;
  sources: string[];
  version: string;
  description: string;
  keywords: string[];
  maintainers: HelmChartMaintainer[];
  engine: string;
  icon: string;
  appVersion: string;
  apiVersion: string;
  urls: string[];
  created: string;
  digest: string;
  labels: Label[];
  deprecated?: boolean;
}

export interface HelmChartDetail {
  metadata: HelmChartMetaData;
  dependencies: HelmChartDependency[];
  values: any;
  files: HelmchartFile;
  security: HelmChartSecurity;
  labels: Label[];
}

export interface HelmChartMetaData {
  name: string;
  home: string;
  sources: string[];
  version: string;
  description: string;
  keywords: string[];
  maintainers: HelmChartMaintainer[];
  engine: string;
  icon: string;
  appVersion: string;
  urls: string[];
  created?: string;
  digest: string;
}

export interface HelmChartMaintainer {
  name: string;
  email: string;
}

export interface HelmChartDependency {
  name: string;
  version: string;
  repository: string;
}

export interface HelmchartFile {
  "README.MD": string;
  "values.yaml": string;
}

export interface HelmChartSecurity {
  signature: HelmChartSignature;
}

export interface HelmChartSignature {
  signed: boolean;
  prov_file: string;
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
