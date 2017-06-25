/**
 * Define the inline template and styles with ts variables
 */

export const LOG_TEMPLATE: string = `
<div>
    <h2 class="h2-log-override" *ngIf="withTitle">{{'SIDE_NAV.LOGS' | translate}}</h2>
    <div class="row flex-items-xs-between flex-items-xs-bottom">
        <div></div>
        <div class="action-head-pos">
            <hbr-filter [withDivider]="true" filterPlaceholder='{{"AUDIT_LOG.FILTER_PLACEHOLDER" | translate}}' (filter)="doFilter($event)" [currentValue]="currentTerm"></hbr-filter>
            <span (click)="refresh()" class="refresh-btn">
            <clr-icon shape="refresh" [hidden]="inProgress" ng-disabled="inProgress"></clr-icon>
            <span class="spinner spinner-inline" [hidden]="!inProgress"></span>
            </span>
        </div>
    </div>
    <div>
        <clr-datagrid (clrDgRefresh)="load($event)" [clrDgLoading]="loading">
            <clr-dg-column [clrDgField]="'username'">{{'AUDIT_LOG.USERNAME' | translate}}</clr-dg-column>
            <clr-dg-column [clrDgField]="'repo_name'">{{'AUDIT_LOG.REPOSITORY_NAME' | translate}}</clr-dg-column>
            <clr-dg-column [clrDgField]="'repo_tag'">{{'AUDIT_LOG.TAGS' | translate}}</clr-dg-column>
            <clr-dg-column [clrDgField]="'operation'">{{'AUDIT_LOG.OPERATION' | translate}}</clr-dg-column>
            <clr-dg-column [clrDgSortBy]="opTimeComparator">{{'AUDIT_LOG.TIMESTAMP' | translate}}</clr-dg-column>
            <clr-dg-placeholder>We couldn't find any logs!</clr-dg-placeholder>
            <clr-dg-row *ngFor="let l of recentLogs">
                <clr-dg-cell>{{l.username}}</clr-dg-cell>
                <clr-dg-cell>{{l.repo_name}}</clr-dg-cell>
                <clr-dg-cell>{{l.repo_tag}}</clr-dg-cell>
                <clr-dg-cell>{{l.operation}}</clr-dg-cell>
                <clr-dg-cell>{{l.op_time | date: 'short'}}</clr-dg-cell>
            </clr-dg-row>
            <clr-dg-footer>
            {{pagination.firstItem + 1}} - {{pagination.lastItem + 1}}
    of {{pagination.totalItems}} {{'AUDIT_LOG.ITEMS' | translate}}
            <clr-dg-pagination #pagination [(clrDgPage)]="currentPage" [clrDgPageSize]="pageSize" [clrDgTotalItems]="totalCount"></clr-dg-pagination>
            </clr-dg-footer>
        </clr-datagrid>
    </div>
</div>
`;

export const LOG_STYLES: string = `
.h2-log-override {
    margin-top: 0px !important;
}

.action-head-pos {
    padding-right: 18px;
    height: 24px;
}

.refresh-btn {
    cursor: pointer;
}

.refresh-btn:hover {
    color: #007CBB;
}

.custom-lines-button {
    padding: 0px !important;
    min-width: 25px !important;
}

.lines-button-toggole {
    font-size: 16px;
    text-decoration: underline;
}

.log-select {
    width: 130px;
    display: inline-block;
    top: 1px;
}

.item-divider {
    height: 24px;
    display: inline-block;
    width: 1px;
    background-color: #ccc;
    opacity: 0.55;
    margin-left: 12px;
    top: 8px;
    position: relative;
}
`;