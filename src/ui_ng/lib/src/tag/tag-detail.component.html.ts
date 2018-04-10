export const TAG_DETAIL_HTML: string = `
<div>
    <section class="overview-section">
        <div class="title-wrapper">
            <div class="arrow-block">
                <a (click)="onBackPro()">< {{'SIDE_NAV.PROJECTS'| translate}}</a>
                <a (click)="onBackRep()">< {{'SIDE_NAV.SYSTEM_MGMT.REGISTRY'| translate}}</a>
                <a (click)="onBackTag()">< {{repositoryId}}</a>
            </div>
            <div class="">
                <h2 class="custom-h2" sub-header-title>{{repositoryId}}:{{tagDetails.name}}</h2>
            </div>
        </div>
        <div class="summary-block">
            <div class="image-summary">
                <div class="flex-block">
                    <div class="image-detail-label">
                        <div>{{'TAG.AUTHOR' | translate }}</div>
                        <div>{{'TAG.ARCHITECTURE' | translate }}</div>
                        <div>{{'TAG.OS' | translate }}</div>
                        <div>{{'TAG.DOCKER_VERSION' | translate }}</div>
                        <div>{{'TAG.SCAN_COMPLETION_TIME' | translate }}</div>
                    </div>
                    <div class="image-detail-value">
                        <div>{{author | translate}}</div>
                        <div>{{tagDetails.architecture}}</div>
                        <div>{{tagDetails.os}}</div>
                        <div>{{tagDetails.docker_version}}</div>
                        <div>{{scanCompletedDatetime | date}}</div>
                    </div>
                </div>
            </div>
            <div *ngIf="withClair">
                <div class="vulnerability">
                    <hbr-vulnerability-bar [repoName]="repositoryId" [tagId]="tagDetails.name" [summary]="tagDetails.scan_overview"></hbr-vulnerability-bar>
                </div>
                <div class="flex-block vulnerabilities-info">
                    <div>
                        <div>
                            <clr-icon shape="error" size="24" class="is-error"></clr-icon>
                        </div>
                        <div class="second-row">
                            <clr-icon shape="exclamation-triangle" size="24" class="tip-icon-medium"></clr-icon>
                        </div>
                        <div>
                            <clr-icon shape="play" size="20" class="tip-icon-low rotate-90"></clr-icon>
                        </div>
                        <div class="second-row">
                            <clr-icon shape="help" size="18" style="margin-left: 2px;"></clr-icon>
                        </div>
                    </div>
                    <div class="second-column">
                        <div>{{highCount}} {{'VULNERABILITY.SEVERITY.HIGH' | translate }}{{'TAG.LEVEL_VULNERABILITIES' | translate }}</div>
                        <div class="second-row">{{mediumCount}} {{'VULNERABILITY.SEVERITY.MEDIUM' | translate }}{{'TAG.LEVEL_VULNERABILITIES' | translate }}</div>
                        <div>{{lowCount}} {{'VULNERABILITY.SEVERITY.LOW' | translate }}{{'TAG.LEVEL_VULNERABILITIES' | translate }}</div>
                        <div class="second-row">{{unknownCount}} {{'VULNERABILITY.SEVERITY.UNKNOWN' | translate }}{{'TAG.LEVEL_VULNERABILITIES' | translate }}</div>
                    </div>
                </div>

            </div>
            <div *ngIf="!withAdmiral && tagDetails?.labels?.length" >
                <div class="third-column detail-title">{{'TAG.LABELS' | translate }}</div>
                <div class="fourth-column">
                  <div *ngFor="let label of tagDetails.labels" style="margin-bottom: 2px;"><hbr-label-piece [label]="label"></hbr-label-piece></div>
                </div>
            </div>
        </div>
    </section>
    <section class="detail-section">
        <div class="vulnerability-block">
            <hbr-vulnerabilities-grid [repositoryId]="repositoryId" [tagId]="tagId"></hbr-vulnerabilities-grid>
        </div>
        <div>
            <ng-content></ng-content>
        </div>
    </section>
</div>
`;