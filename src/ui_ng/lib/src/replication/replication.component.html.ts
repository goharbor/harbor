export const REPLICATION_TEMPLATE: string = `
<div class="row">
  <div class="col-lg-12 col-md-12 col-sm-12 col-xs-12">
    <div class="row flex-items-xs-between" style="height:32px;">
      <div class="flex-xs-middle option-left">
        <button *ngIf="projectId" class="btn btn-link" (click)="openModal()"><clr-icon shape="add"></clr-icon> {{'REPLICATION.REPLICATION_RULE' | translate}}</button>
        <create-edit-rule [projectId]="projectId" (reload)="reloadRules($event)"></create-edit-rule>
      </div>
      <div class="flex-xs-middle option-right">
        <div class="select" style="float: left; top: 8px;">
          <select (change)="doFilterRuleStatus($event)">
            <option *ngFor="let r of ruleStatus" value="{{r.key}}">{{r.description | translate}}</option>
          </select>
        </div> 
        <hbr-filter [withDivider]="true" filterPlaceholder='{{"REPLICATION.FILTER_POLICIES_PLACEHOLDER" | translate}}' (filter)="doSearchRules($event)" [currentValue]="search.ruleName"></hbr-filter>
        <span class="refresh-btn" (click)="refreshRules()">
          <clr-icon shape="refresh"></clr-icon>
        </span>
      </div>
    </div>
    </div>
    <div class="col-lg-12 col-md-12 col-sm-12 col-xs-12">
      <hbr-list-replication-rule #listReplicationRule [projectId]="projectId" (selectOne)="selectOneRule($event)" (editOne)="openEditRule($event)" (reload)="reloadRules($event)" [loading]="loading" [withReplicationJob]="withReplicationJob" (redirect)="customRedirect($event)"></hbr-list-replication-rule>
    </div>
    <div *ngIf="withReplicationJob" class="col-lg-12 col-md-12 col-sm-12 col-xs-12">
      <div class="row flex-items-xs-between" style="height:60px;">
        <h5 class="flex-items-xs-bottom option-left-down" style="margin-left: 14px;">{{'REPLICATION.REPLICATION_JOBS' | translate}}</h5>
        <div class="flex-items-xs-bottom option-right-down">
          <button class="btn btn-link" (click)="toggleSearchJobOptionalName(currentJobSearchOption)">{{toggleJobSearchOption[currentJobSearchOption] | translate}}</button>
          <hbr-filter [withDivider]="true" filterPlaceholder='{{"REPLICATION.FILTER_JOBS_PLACEHOLDER" | translate}}' (filter)="doSearchJobs($event)" [currentValue]="search.repoName" ></hbr-filter>
          <span class="refresh-btn" (click)="refreshJobs()">
            <clr-icon shape="refresh"></clr-icon>
          </span>
        </div>
      </div>
      <div class="row flex-items-xs-right option-right" [hidden]="currentJobSearchOption === 0">
        <div class="select" style="float: left;">
          <select (change)="doFilterJobStatus($event)">
          <option *ngFor="let j of jobStatus" value="{{j.key}}" [selected]="currentJobStatus.key === j.key">{{j.description | translate}}</option>
          </select>
        </div>
        <div class="flex-items-xs-middle">    
          <hbr-datetime [dateInput]="search.startTime" (search)="doJobSearchByStartTime($event)"></hbr-datetime>
          <hbr-datetime [dateInput]="search.endTime" [oneDayOffset]="true" (search)="doJobSearchByEndTime($event)"></hbr-datetime>
        </div>
      </div>
    </div>
    <div *ngIf="withReplicationJob" class="col-lg-12 col-md-12 col-sm-12 col-xs-12">
      <clr-datagrid [clrDgLoading]="loading">
        <clr-dg-column [clrDgField]="'repository'">{{'REPLICATION.NAME' | translate}}</clr-dg-column>
        <clr-dg-column [clrDgField]="'status'">{{'REPLICATION.STATUS' | translate}}</clr-dg-column>
        <clr-dg-column [clrDgField]="'operation'">{{'REPLICATION.OPERATION' | translate}}</clr-dg-column>
        <clr-dg-column [clrDgSortBy]="creationTimeComparator">{{'REPLICATION.CREATION_TIME' | translate}}</clr-dg-column>
        <clr-dg-column [clrDgSortBy]="updateTimeComparator">{{'REPLICATION.END_TIME' | translate}}</clr-dg-column>
        <clr-dg-column>{{'REPLICATION.LOGS' | translate}}</clr-dg-column>
        <clr-dg-placeholder>{{'REPLICATION.JOB_PLACEHOLDER' | translate }}</clr-dg-placeholder>
        <clr-dg-row *clrDgItems="let j of jobs" [clrDgItem]='j'>
            <clr-dg-cell>{{j.repository}}</clr-dg-cell>
            <clr-dg-cell>{{j.status}}</clr-dg-cell>
            <clr-dg-cell>{{j.operation}}</clr-dg-cell>
            <clr-dg-cell>{{j.creation_time | date: 'short'}}</clr-dg-cell>
            <clr-dg-cell>{{j.update_time | date: 'short'}}</clr-dg-cell>
            <clr-dg-cell>
              <a href="/api/jobs/replication/{{j.id}}/log" target="_BLANK">
                <clr-icon shape="clipboard"></clr-icon>
              </a>
            </clr-dg-cell>
        </clr-dg-row>
        <clr-dg-footer>
            <span *ngIf="pagination.totalItems">{{pagination.firstItem + 1}} - {{pagination.lastItem + 1}} {{'REPLICATION.OF' | translate}}</span>
            {{pagination.totalItems}} {{'REPLICATION.ITEMS' | translate}}
            <clr-dg-pagination #pagination [clrDgPageSize]="5"></clr-dg-pagination>
        </clr-dg-footer>
      </clr-datagrid>
    </div>
</div>`;