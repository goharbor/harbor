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
          <hbr-datetime [dateInput]="search.startTime" (search)="doJobSearchByStartTime($event)"></hbr-datetime>
          <hbr-datetime [dateInput]="search.endTime" [oneDayOffset]="true" (search)="doJobSearchByEndTime($event)"></hbr-datetime>
        </div>
      </div>
    </div>
    <div class="col-lg-12 col-md-12 col-sm-12 col-xs-12">
      <list-replication-job [jobs]="changedJobs" [totalPage]="jobsTotalPage" [totalRecordCount]="jobsTotalRecordCount" (paginate)="fetchReplicationJobs($event)"></list-replication-job>
    </div>
</div>`;