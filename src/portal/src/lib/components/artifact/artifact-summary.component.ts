import { Component, Input, Output, EventEmitter, OnInit } from "@angular/core";

import { TagService, Tag, VulnerabilitySeverity, VulnerabilitySummary, ArtifactService, ProjectService } from "../../services";
import { ErrorHandler } from "../../utils/error-handler";
import { Label } from "../../services/interface";
import { forkJoin } from "rxjs";
import { UserPermissionService } from "../../services/permission.service";
import { USERSTATICPERMISSION } from "../../services/permission-static";
import { ChannelService } from "../../services/channel.service";
import { DEFAULT_SUPPORTED_MIME_TYPE, VULNERABILITY_SCAN_STATUS, VULNERABILITY_SEVERITY } from "../../utils/utils";
import { Reference, Artifact } from "./artifact";

const TabLinkContentMap: { [index: string]: string } = {
  "tag-history": "history",
  "tag-vulnerability": "vulnerability"
};

@Component({
  selector: "artifact-summary",
  templateUrl: "./artifact-summary.component.html",
  styleUrls: ["./artifact-summary.component.scss"],

  providers: []
})
export class ArtifactSummaryComponent implements OnInit {
  _highCount: number = 0;
  _mediumCount: number = 0;
  _lowCount: number = 0;
  _unknownCount: number = 0;
  labels: Label;
  vulnerabilitySummary: VulnerabilitySummary;
  @Input()
  artifactDigest: string;
  @Input()
  repositoryName: string;
  @Input()
  withAdmiral: boolean;
  artifactDetails: Artifact;
  //  =
  //   {
  //     "id": 1,
  //     type: 'image',
  //     repository: "goharbor/harbor-portal",
  //     tags: [{
  //       id: '1',
  //       name: 'tag1',
  //       upload_time: '2020-01-06T09:40:08.036866579Z',
  //       latest_download_time: '2020-01-06T09:40:08.036866579Z',
  //   },
  //   {
  //       id: '2',
  //       name: 'tag2',
  //       upload_time: '2020-01-06T09:40:08.036866579Z',
  //       latest_download_time: '2020-01-06T09:40:08.036866579Z',
  //   },],
  //     references: [new Reference(1), new Reference(2)],
  //     media_type: 'string',
  //     "digest": "sha256:4875cda368906fd670c9629b5e416ab3d6c0292015f3c3f12ef37dc9a32fc8d4",
  //     "size": 20372934,
  //     "scan_overview": {
  //       "application/vnd.scanner.adapter.vuln.report.harbor+json; version=1.0": {
  //         "report_id": "5e64bc05-3102-11ea-93ae-0242ac140004",
  //         "scan_status": "Error",
  //         "severity": "",
  //         "duration": 118,
  //         "summary": null,
  //         "start_time": "2020-01-07T04:01:23.157711Z",
  //         "end_time": "2020-01-07T04:03:21.662766Z"
  //       }
  //     },
  //     "labels": [
  //       {
  //         "id": 3,
  //         "name": "aaa",
  //         "description": "",
  //         "color": "#0095D3",
  //         "scope": "g",
  //         "project_id": 0,
  //         "creation_time": "2020-01-13T05:44:00.580198Z",
  //         "update_time": "2020-01-13T05:44:00.580198Z",
  //         "deleted": false
  //       },
  //       {
  //         "id": 6,
  //         "name": "dbc",
  //         "description": "",
  //         "color": "",
  //         "scope": "g",
  //         "project_id": 0,
  //         "creation_time": "2020-01-13T08:27:19.279123Z",
  //         "update_time": "2020-01-13T08:27:19.279123Z",
  //         "deleted": false
  //       }
  //     ],
  //     "push_time": "2020-01-07T03:33:41.162319Z",
  //     "pull_time": "0001-01-01T00:00:00Z",
  //     hasReferenceArtifactList: [],
  //     noReferenceArtifactList: []

  // };
  // artifactDetails1: Artifact = this.artifactDetails;
  @Output()
  backEvt: EventEmitter<any> = new EventEmitter<any>();

  currentTabID = "tag-vulnerability";
  hasVulnerabilitiesListPermission: boolean;
  hasBuildHistoryPermission: boolean;
  @Input() projectId: number;
  projectName: string;
  showStatBar: boolean = true;

  constructor(
    private projectService: ProjectService,
    private artifactService: ArtifactService,
    public channel: ChannelService,
    private errorHandler: ErrorHandler,
    private userPermissionService: UserPermissionService
  ) { }

  ngOnInit(): void {
    if (this.repositoryName && this.artifactDigest) {
      // this.tagService.getTag(this.repositoryId, this.tagId).subscribe(
        this.projectService.getProject(this.projectId).subscribe(project => {
          this.projectName = project.name;
          this.getArtifact();
        })
      
    }
    this.getTagPermissions(this.projectId);
    this.channel.tagDetail$.subscribe(artifact => {
      this.getArtifactDetails(artifact);
    });
  }
  getArtifact() {
    this.artifactService.getArtifactFromId(this.projectName, this.repositoryName, this.artifactDigest).subscribe(
      response => {
        this.getArtifactDetails(response);
      },
      error => this.errorHandler.error(error)
    );
  }
  getArtifactDetails(artifactDetails: Artifact): void {
    this.artifactDetails = artifactDetails;
    // || this.artifactDetails1;
    if (this.artifactDetails
      && this.artifactDetails.scan_overview
      && this.artifactDetails.scan_overview[DEFAULT_SUPPORTED_MIME_TYPE]) {
      this.vulnerabilitySummary = this.artifactDetails.scan_overview[DEFAULT_SUPPORTED_MIME_TYPE];
      this.showStatBar = false;
    }
  }
  onBack(): void {
    this.backEvt.emit(this.repositoryName);
  }

  getPackageText(count: number): string {
    return count > 1 ? "VULNERABILITY.PACKAGES" : "VULNERABILITY.PACKAGE";
  }

  packageText(count: number): string {
    return count > 1
      ? "VULNERABILITY.GRID.COLUMN_PACKAGES"
      : "VULNERABILITY.GRID.COLUMN_PACKAGE";
  }

  haveText(count: number): string {
    return count > 1 ? "TAG.HAVE" : "TAG.HAS";
  }

  // public get author(): string {
  //   return this.artifactDetails && this.artifactDetails.author
  //     ? this.artifactDetails.author
  //     : "TAG.ANONYMITY";
  // }
  private getCountByLevel(level: string): number {
    if (this.vulnerabilitySummary && this.vulnerabilitySummary.summary
      && this.vulnerabilitySummary.summary.summary) {
      return this.vulnerabilitySummary.summary.summary[level];
    }
    return 0;
  }
  /**
   *  count of critical level vulnerabilities
   */
  get criticalCount(): number {
    return this.getCountByLevel(VULNERABILITY_SEVERITY.CRITICAL);
  }

  /**
   *  count of high level vulnerabilities
   */
  get highCount(): number {
    return this.getCountByLevel(VULNERABILITY_SEVERITY.HIGH);
  }
  /**
   *  count of medium level vulnerabilities
   */
  get mediumCount(): number {
    return this.getCountByLevel(VULNERABILITY_SEVERITY.MEDIUM);
  }
  /**
   *  count of low level vulnerabilities
   */
  get lowCount(): number {
    return this.getCountByLevel(VULNERABILITY_SEVERITY.LOW);
  }
  /**
   *  count of unknown vulnerabilities
   */
  get unknownCount(): number {
    return this.getCountByLevel(VULNERABILITY_SEVERITY.UNKNOWN);
  }
  /**
   *  count of negligible vulnerabilities
   */
  get negligibleCount(): number {
    return this.getCountByLevel(VULNERABILITY_SEVERITY.NEGLIGIBLE);
  }
  get hasCve(): boolean {
    return this.vulnerabilitySummary
      && this.vulnerabilitySummary.scan_status === VULNERABILITY_SCAN_STATUS.SUCCESS
      && this.vulnerabilitySummary.severity !== VULNERABILITY_SEVERITY.NONE;
  }
  public get scanCompletedDatetime(): Date {
    return this.artifactDetails && this.artifactDetails.scan_overview
      && this.artifactDetails.scan_overview[DEFAULT_SUPPORTED_MIME_TYPE]
      ? this.artifactDetails.scan_overview[DEFAULT_SUPPORTED_MIME_TYPE].end_time
      : null;
  }

  public get suffixForHigh(): string {
    return this.highCount > 1
      ? "VULNERABILITY.PLURAL"
      : "VULNERABILITY.SINGULAR";
  }

  public get suffixForMedium(): string {
    return this.mediumCount > 1
      ? "VULNERABILITY.PLURAL"
      : "VULNERABILITY.SINGULAR";
  }

  public get suffixForLow(): string {
    return this.lowCount > 1
      ? "VULNERABILITY.PLURAL"
      : "VULNERABILITY.SINGULAR";
  }

  public get suffixForUnknown(): string {
    return this.unknownCount > 1
      ? "VULNERABILITY.PLURAL"
      : "VULNERABILITY.SINGULAR";
  }

  isCurrentTabLink(tabID: string): boolean {
    return this.currentTabID === tabID;
  }

  isCurrentTabContent(ContentID: string): boolean {
    return TabLinkContentMap[this.currentTabID] === ContentID;
  }

  tabLinkClick(tabID: string) {
    this.currentTabID = tabID;
  }

  getTagPermissions(projectId: number): void {
    const hasVulnerabilitiesListPermission = this.userPermissionService.getPermission(
      projectId,
      USERSTATICPERMISSION.REPOSITORY_TAG_VULNERABILITY.KEY,
      USERSTATICPERMISSION.REPOSITORY_TAG_VULNERABILITY.VALUE.LIST
    );
    const hasBuildHistoryPermission = this.userPermissionService.getPermission(
      projectId,
      USERSTATICPERMISSION.REPOSITORY_TAG_MANIFEST.KEY,
      USERSTATICPERMISSION.REPOSITORY_TAG_MANIFEST.VALUE.READ
    );
    forkJoin(
      hasVulnerabilitiesListPermission,
      hasBuildHistoryPermission
    ).subscribe(
      permissions => {
        this.hasVulnerabilitiesListPermission = permissions[0] as boolean;
        this.hasBuildHistoryPermission = permissions[1] as boolean;
      },
      error => this.errorHandler.error(error)
    );
  }
  passMetadataToChart() {
    return [
      {
        text: 'VULNERABILITY.SEVERITY.CRITICAL',
        value: this.criticalCount ? this.criticalCount : 0,
        color: 'red'
      },
      {
        text: 'VULNERABILITY.SEVERITY.HIGH',
        value: this.highCount ? this.highCount : 0,
        color: '#e64524'
      },
      {
        text: 'VULNERABILITY.SEVERITY.MEDIUM',
        value: this.mediumCount ? this.mediumCount : 0,
        color: 'orange'
      },
      {
        text: 'VULNERABILITY.SEVERITY.LOW',
        value: this.lowCount ? this.lowCount : 0,
        color: '#007CBB'
      },
      {
        text: 'VULNERABILITY.SEVERITY.NEGLIGIBLE',
        value: this.negligibleCount ? this.negligibleCount : 0,
        color: 'green'
      },
      {
        text: 'VULNERABILITY.SEVERITY.UNKNOWN',
        value: this.unknownCount ? this.unknownCount : 0,
        color: 'grey'
      },
    ];
  }
  isThemeLight() {
    return localStorage.getItem('styleModeLocal') === 'LIGHT';
  }
  refreshArtifact() {
    this.getArtifact();
  }
}
