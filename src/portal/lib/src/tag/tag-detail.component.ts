import { Component, Input, Output, EventEmitter, OnInit } from "@angular/core";

import { TagService, Tag, VulnerabilitySeverity } from "../service/index";
import { ErrorHandler } from "../error-handler/index";
import { Label } from "../service/interface";
import { forkJoin } from "rxjs";
import { UserPermissionService } from "../service/permission.service";
import { USERSTATICPERMISSION } from "../service/permission-static";
import { ChannelService } from "../channel/channel.service";

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
  getTagDetails(tagDetails): void {
    this._highCount = 0;
    this._mediumCount = 0;
    this._lowCount = 0;
    this._unknownCount = 0;
    this.tagDetails = tagDetails;
    if (
      this.tagDetails &&
      this.tagDetails.scan_overview &&
      this.tagDetails.scan_overview.components &&
      this.tagDetails.scan_overview.components.summary
    ) {
      this.tagDetails.scan_overview.components.summary.forEach(item => {
        switch (item.severity) {
          case VulnerabilitySeverity.UNKNOWN:
            this._unknownCount += item.count;
            break;
          case VulnerabilitySeverity.LOW:
            this._lowCount += item.count;
            break;
          case VulnerabilitySeverity.MEDIUM:
            this._mediumCount += item.count;
            break;
          case VulnerabilitySeverity.HIGH:
            this._highCount += item.count;
            break;
          default:
            break;
        }
      });
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

  public get highCount(): number {
    return this._highCount;
  }

  public get mediumCount(): number {
    return this._mediumCount;
  }

  public get lowCount(): number {
    return this._lowCount;
  }

  public get unknownCount(): number {
    return this._unknownCount;
  }

  public get scanCompletedDatetime(): Date {
    return this.tagDetails && this.tagDetails.scan_overview
      ? this.tagDetails.scan_overview.update_time
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
