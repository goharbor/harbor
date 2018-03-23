export const TAG_DETAIL_HTML: string = `
<div>
    <section class="overview-section">
        <div class="title-wrapper">
            <div class="title-block arrow-block">
                <clr-icon class="rotate-90 arrow-back" shape="arrow" size="36" (click)="onBack()"></clr-icon>
            </div>
            <div class="title-block">
                <div class="tag-name">
                    <h1>{{tagDetails.name}}</h1>
                </div>
                <div class="tag-timestamp">
                    {{'TAG.CREATION_TIME_PREFIX' | translate }} {{tagDetails.created | date }} {{'TAG.CREATOR_PREFIX' | translate }} {{author | translate}}
                </div>
            </div>
        </div>
        <div class="summary-block">
            <div class="image-summary">
                <div class="detail-title">
                    {{'TAG.IMAGE_DETAILS' | translate }}
                </div>
                <div class="flex-block">
                    <div class="image-detail-label">
                        <div>{{'TAG.ARCHITECTURE' | translate }}</div>
                        <div>{{'TAG.OS' | translate }}</div>
                        <div>{{'TAG.DOCKER_VERSION' | translate }}</div>
                        <div>{{'TAG.SCAN_COMPLETION_TIME' | translate }}</div>
                    </div>
                    <div class="image-detail-value">
                        <div>{{tagDetails.architecture}}</div>
                        <div>{{tagDetails.os}}</div>
                        <div>{{tagDetails.docker_version}}</div>
                        <div>{{scanCompletedDatetime | date}}</div>
                    </div>
                </div>
            </div>
            <div>
                <div class="detail-title">
                    {{'TAG.IMAGE_VULNERABILITIES' | translate }}
                </div>
                <div class="flex-block vulnerabilities-info">
                    <div>
                        <div>
                            <clr-icon shape="error" size="24" class="is-error"></clr-icon>
                        </div>
                        <div class="second-row">
                            <clr-icon shape="exclamation-triangle" size="24" class="tip-icon-medium"></clr-icon>
                        </div>
                    </div>
                    <div class="second-column">
                        <div>{{highCount}} {{'VULNERABILITY.SEVERITY.HIGH' | translate }}</div>
                        <div class="second-row">{{mediumCount}} {{'VULNERABILITY.SEVERITY.MEDIUM' | translate }}</div>
                    </div>
                    <div class="third-column">
                        <div>
                            <clr-icon shape="play" size="20" class="tip-icon-low rotate-90"></clr-icon>
                        </div>
                        <div class="second-row">
                            <clr-icon shape="help" size="18" style="margin-left: 2px;"></clr-icon>
                        </div>
                    </div>
                    <div class="fourth-column">
                        <div>{{lowCount}} {{'VULNERABILITY.SEVERITY.LOW' | translate }}</div>
                        <div class="second-row">{{unknownCount}} {{'VULNERABILITY.SEVERITY.UNKNOWN' | translate }}</div>
                    </div>
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