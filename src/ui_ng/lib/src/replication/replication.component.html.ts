export const REPLICATION_TEMPLATE: string = `
<div class="row">
  <div class="col-lg-12 col-md-12 col-sm-12 col-xs-12">
    <div class="row flex-items-xs-between">
      <div class="flex-xs-middle option-left">
        <button class="btn btn-link" (click)="openModal()"><clr-icon shape="add"></clr-icon> {{'REPLICATION.REPLICATION_RULE' | translate}}</button>
        <create-edit-rule [projectId]="projectId" (reload)="reloadRules($event)"></create-edit-rule>
      </div>
      <div class="flex-xs-middle option-right">
        <div class="select" style="float: left;">
          <select (change)="doFilterRuleStatus($event)">
            <option *ngFor="let r of ruleStatus" value="{{r.key}}">{{r.description | translate}}</option>
          </select>
        </div> 
        <hbr-filter filterPlaceholder='{{"REPLICATION.FILTER_POLICIES_PLACEHOLDER" | translate}}' (filter)="doSearchRules($event)" [currentValue]="search.ruleName"></hbr-filter>
        <a href="javascript:void(0)" (click)="refreshRules()">
          <clr-icon shape="refresh"></clr-icon>
        </a>
      </div>
    </div>
    </div>
    <div class="col-lg-12 col-md-12 col-sm-12 col-xs-12">
      <list-replication-rule [rules]="changedRules" [projectless]="false" [selectedId]="initSelectedId" (selectOne)="selectOneRule($event)" (editOne)="openEditRule($event)" (reload)="reloadRules($event)"></list-replication-rule>
    </div>
    <div class="col-lg-12 col-md-12 col-sm-12 col-xs-12">
      <div class="row flex-items-xs-between">
        <h5 class="flex-items-xs-bottom option-left-down" style="margin-left: 14px;">{{'REPLICATION.REPLICATION_JOBS' | translate}}</h5>
        <div class="flex-items-xs-bottom option-right-down">
          <button class="btn btn-link" (click)="toggleSearchJobOptionalName(currentJobSearchOption)">{{toggleJobSearchOption[currentJobSearchOption] | translate}}</button>
          <hbr-filter filterPlaceholder='{{"REPLICATION.FILTER_JOBS_PLACEHOLDER" | translate}}' (filter)="doSearchJobs($event)" [currentValue]="search.repoName" ></hbr-filter>
          <a href="javascript:void(0)" (click)="refreshJobs()">
            <clr-icon shape="refresh"></clr-icon>
          </a>
        </div>
      </div>
      <div class="row flex-items-xs-right option-right" [hidden]="currentJobSearchOption === 0">
        <div class="select" style="float: left;">
          <select (change)="doFilterJobStatus($event)">
          <option *ngFor="let j of jobStatus" value="{{j.key}}" [selected]="currentJobStatus.key === j.key">{{j.description | translate}}</option>
          </select>
        </div>
        <div class="flex-items-xs-middle">
          <clr-icon shape="date"></clr-icon>
          <label for="fromDateInput" aria-haspopup="true" role="tooltip" [class.invalid]="fromTimeInvalid" class="tooltip tooltip-validation invalid tooltip-sm">
            <input id="fromDateInput" type="date" #fromTime="ngModel" name="from" [(ngModel)]="search.startTime" dateValidator placeholder="dd/mm/yyyy" (change)="doJobSearchByStartTime(fromTime.value)">
            <span *ngIf="fromTimeInvalid" class="tooltip-content">
               {{'AUDIT_LOG.INVALID_DATE' | translate }}
            </span>
          </label>
          <clr-icon shape="date"></clr-icon>
          <label for="toDateInput" aria-haspopup="true" role="tooltip" [class.invalid]="toTimeInvalid" class="tooltip tooltip-validation invalid tooltip-sm">
            <input id="toDateInput" type="date" #toTime="ngModel" name="to" [(ngModel)]="search.endTime" dateValidator placeholder="dd/mm/yyyy" (change)="doJobSearchByEndTime(toTime.value)">
            <span *ngIf="toTimeInvalid" class="tooltip-content">
               {{'AUDIT_LOG.INVALID_DATE' | translate }}
            </span>
          </label>
        </div>
      </div>
    </div>
    <div class="col-lg-12 col-md-12 col-sm-12 col-xs-12">
      <list-replication-job [jobs]="changedJobs" [totalPage]="jobsTotalPage" [totalRecordCount]="jobsTotalRecordCount" (paginate)="fetchReplicationJobs($event)"></list-replication-job>
    </div>
</div>`;