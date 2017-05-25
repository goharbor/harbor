/**
 * Define the inline template and styles with ts variables
 */

export const LOG_TEMPLATE: string = `
<div>
    <h2 class="h2-log-override">{{'SIDE_NAV.LOGS' | translate}}</h2>
    <div class="row flex-items-xs-between flex-items-xs-bottom">
        <div></div>
        <div class="action-head-pos">
            <div class="select log-select">
                <select id="log_display_num" (change)="handleOnchange($event)">
                    <option value="10">{{'RECENT_LOG.SUB_TITLE' | translate}} 10 {{'RECENT_LOG.SUB_TITLE_SUFIX' | translate}}</option>
                    <option value="25">{{'RECENT_LOG.SUB_TITLE' | translate}} 25 {{'RECENT_LOG.SUB_TITLE_SUFIX' | translate}}</option>
                    <option value="50">{{'RECENT_LOG.SUB_TITLE' | translate}} 50 {{'RECENT_LOG.SUB_TITLE_SUFIX' | translate}}</option>
            </select>
            </div>
            <div class="item-divider"></div>
            <hbr-filter filterPlaceholder='{{"AUDIT_LOG.FILTER_PLACEHOLDER" | translate}}' (filter)="doFilter($event)" [currentValue]="currentTerm"></hbr-filter>
            <span (click)="refresh()" class="refresh-btn">
            <clr-icon shape="refresh" [hidden]="inProgress" ng-disabled="inProgress"></clr-icon>
            <span class="spinner spinner-inline" [hidden]="inProgress === false"></span>
            </span>
        </div>
    </div>
    <div>
        <clr-datagrid [clrDgLoading]="loading">
            <clr-dg-column [clrDgField]="'username'">{{'AUDIT_LOG.USERNAME' | translate}}</clr-dg-column>
            <clr-dg-column [clrDgField]="'repo_name'">{{'AUDIT_LOG.REPOSITORY_NAME' | translate}}</clr-dg-column>
            <clr-dg-column [clrDgField]="'repo_tag'">{{'AUDIT_LOG.TAGS' | translate}}</clr-dg-column>
            <clr-dg-column [clrDgField]="'operation'">{{'AUDIT_LOG.OPERATION' | translate}}</clr-dg-column>
            <clr-dg-column [clrDgSortBy]="opTimeComparator">{{'AUDIT_LOG.TIMESTAMP' | translate}}</clr-dg-column>
            <clr-dg-row *clrDgItems="let l of recentLogs">
                <clr-dg-cell>{{l.username}}</clr-dg-cell>
                <clr-dg-cell>{{l.repo_name}}</clr-dg-cell>
                <clr-dg-cell>{{l.repo_tag}}</clr-dg-cell>
                <clr-dg-cell>{{l.operation}}</clr-dg-cell>
                <clr-dg-cell>{{l.op_time | date: 'short'}}</clr-dg-cell>
            </clr-dg-row>
            <clr-dg-footer>{{ (recentLogs ? recentLogs.length : 0) }} {{'AUDIT_LOG.ITEMS' | translate}}</clr-dg-footer>
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
}

.refresh-btn {
    cursor: pointer;
}

.refresh-btn:hover {
    color: #00bfff;
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