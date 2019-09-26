import { Component, Input, Output, EventEmitter, OnInit } from "@angular/core";

import { TagService, Tag, VulnerabilitySeverity, VulnerabilitySummary } from "../service/index";
import { ErrorHandler } from "../error-handler/index";
import { Label } from "../service/interface";
import { forkJoin } from "rxjs";
import { UserPermissionService } from "../service/permission.service";
import { USERSTATICPERMISSION } from "../service/permission-static";
import { ChannelService } from "../channel/channel.service";
import { VULNERABILITY_SCAN_STATUS, VULNERABILITY_SEVERITY } from "../utils";

const TabLinkContentMap: { [index: string]: string } = {
  "tag-history": "history",
  "tag-vulnerability": "vulnerability"
};

@Component({
  selector: "hbr-tag-detail",
  templateUrl: "./tag-detail.component.html",
  styleUrls: ["./tag-detail.component.scss"],

  providers: []
})
export class TagDetailComponent implements OnInit {
  _highCount: number = 0;
  _mediumCount: number = 0;
  _lowCount: number = 0;
  _unknownCount: number = 0;
  labels: Label;
  vulnerabilitySummary: VulnerabilitySummary;
  @Input()
  tagId: string;
  @Input()
  repositoryId: string;
  @Input()
  withAdmiral: boolean;
  @Input()
  withClair: boolean;
  tagDetails: Tag = {
    name: "--",
    size: "--",
    author: "--",
    created: new Date(),
    architecture: "--",
    os: "--",
    "os.version": "--",
    docker_version: "--",
    digest: "--",
    labels: []
  };

  @Output()
  backEvt: EventEmitter<any> = new EventEmitter<any>();

  currentTabID = "tag-vulnerability";
  hasVulnerabilitiesListPermission: boolean;
  hasBuildHistoryPermission: boolean;
  @Input() projectId: number;
  constructor(
    private tagService: TagService,
    public channel: ChannelService,
    private errorHandler: ErrorHandler,
    private userPermissionService: UserPermissionService
  ) {}

  ngOnInit(): void {
    if (this.repositoryId && this.tagId) {
      this.tagService.getTag(this.repositoryId, this.tagId).subscribe(
        response => {
          this.getTagDetails(response);
        },
        error => this.errorHandler.error(error)
      );
    }
    this.getTagPermissions(this.projectId);
    this.channel.tagDetail$.subscribe(tag => {
       this.getTagDetails(tag);
    });
  }
  getTagDetails(tagDetails: Tag): void {
    this.tagDetails = tagDetails;
    if (tagDetails
        && tagDetails.scan_overview
        && tagDetails.scan_overview["application/vnd.scanner.adapter.vuln.report.harbor+json; version=1.0"]) {
      this.vulnerabilitySummary = tagDetails.scan_overview["application/vnd.scanner.adapter.vuln.report.harbor+json; version=1.0"];
    }
  }
  onBack(): void {
    this.backEvt.emit(this.repositoryId);
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

  public get author(): string {
    return this.tagDetails && this.tagDetails.author
      ? this.tagDetails.author
      : "TAG.ANONYMITY";
  }

  get highCount(): number {
    if (this.vulnerabilitySummary && this.vulnerabilitySummary.summary
        && this.vulnerabilitySummary.summary.summary) {
      return this.vulnerabilitySummary.summary.summary[VULNERABILITY_SEVERITY.HIGH];
    }
    return 0;
  }

  get mediumCount(): number {
    if (this.vulnerabilitySummary && this.vulnerabilitySummary.summary
        && this.vulnerabilitySummary.summary.summary) {
      return this.vulnerabilitySummary.summary.summary[VULNERABILITY_SEVERITY.MEDIUM];
    }
    return 0;
  }

  get lowCount(): number {
    if (this.vulnerabilitySummary && this.vulnerabilitySummary.summary
        && this.vulnerabilitySummary.summary.summary) {
      return this.vulnerabilitySummary.summary.summary[VULNERABILITY_SEVERITY.LOW];
    }
    return 0;
  }

  get unknownCount(): number {
    if (this.vulnerabilitySummary && this.vulnerabilitySummary.summary
        && this.vulnerabilitySummary.summary.summary) {
      return this.vulnerabilitySummary.summary.summary[VULNERABILITY_SEVERITY.UNKNOWN];
    }
    return 0;
  }
  get negligibleCount(): number {
    if (this.vulnerabilitySummary && this.vulnerabilitySummary.summary
        && this.vulnerabilitySummary.summary.summary) {
      return this.vulnerabilitySummary.summary.summary[VULNERABILITY_SEVERITY.NEGLIGIBLE];
    }
    return 0;
  }
  get hasCve(): boolean {
    return this.vulnerabilitySummary
           && this.vulnerabilitySummary.scan_status === VULNERABILITY_SCAN_STATUS.SUCCESS;
  }
  public get scanCompletedDatetime(): Date {
    return this.tagDetails && this.tagDetails.scan_overview
    && this.tagDetails.scan_overview["application/vnd.scanner.adapter.vuln.report.harbor+json; version=1.0"]
      ? this.tagDetails.scan_overview["application/vnd.scanner.adapter.vuln.report.harbor+json; version=1.0"].end_time
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
}
