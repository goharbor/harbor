export const REPOSITORY_STACKVIEW_TEMPLATE: string = `
<div>
<div class="row">
  <div class="col-lg-12 col-md-12 col-sm-12 col-xs-12" style="height: 32px;">  
    <div class="row flex-items-xs-right option-right">
      <div class="flex-xs-middle">
        <hbr-push-image-button style="display: inline-block;" [registryUrl]="registryUrl" [projectName]="projectName"></hbr-push-image-button>
        <hbr-filter [withDivider]="true" filterPlaceholder="{{'REPOSITORY.FILTER_FOR_REPOSITORIES' | translate}}" (filter)="doSearchRepoNames($event)" [currentValue]="lastFilteredRepoName"></hbr-filter>  
        <span class="refresh-btn" (click)="refresh()"><clr-icon shape="refresh"></clr-icon></span>
      </div>
    </div>
  </div>
  <div class="col-lg-12 col-md-12 col-sm-12 col-xs-12">  
    <clr-datagrid (clrDgRefresh)="clrLoad($event)" [clrDgLoading]="loading">
      <clr-dg-column [clrDgField]="'name'">{{'REPOSITORY.NAME' | translate}}</clr-dg-column>
      <clr-dg-column [clrDgSortBy]="tagsCountComparator">{{'REPOSITORY.TAGS_COUNT' | translate}}</clr-dg-column>
      <clr-dg-column [clrDgSortBy]="pullCountComparator">{{'REPOSITORY.PULL_COUNT' | translate}}</clr-dg-column>
      <clr-dg-placeholder>{{'REPOSITORY.PLACEHOLDER' | translate}}</clr-dg-placeholder>
      <clr-dg-row *ngFor="let r of repositories">
        <clr-dg-action-overflow [hidden]="!hasProjectAdminRole">
          <button class="action-item" (click)="deleteRepo(r.name)">{{'REPOSITORY.DELETE' | translate}}</button>
        </clr-dg-action-overflow>
        <clr-dg-cell>{{r.name}}</clr-dg-cell>
        <clr-dg-cell>{{r.tags_count}}</clr-dg-cell>
        <clr-dg-cell>{{r.pull_count}}</clr-dg-cell>        
        <hbr-tag *clrIfExpanded ngProjectAs="clr-dg-row-detail" (tagClickEvent)="watchTagClickEvt($event)" class="sub-grid-custom" [repoName]="r.name" [registryUrl]="registryUrl" [withNotary]="withNotary" [withClair]="withClair" [hasSignedIn]="hasSignedIn" [hasProjectAdminRole]="hasProjectAdminRole" [projectId]="projectId" [isEmbedded]="true"></hbr-tag>
      </clr-dg-row>
      <clr-dg-footer>
        <span *ngIf="showDBStatusWarning" class="db-status-warning">
          <clr-icon shape="warning" class="is-warning" size="24"></clr-icon>
          {{'CONFIG.SCANNING.DB_NOT_READY' | translate }}
        </span>
        <span *ngIf="pagination.totalItems">{{pagination.firstItem + 1}} - {{pagination.lastItem + 1}} {{'REPOSITORY.OF' | translate}}</span>
        {{pagination.totalItems}} {{'REPOSITORY.ITEMS' | translate}}
        <clr-dg-pagination #pagination [(clrDgPage)]="currentPage" [clrDgPageSize]="pageSize" [clrDgTotalItems]="totalCount"></clr-dg-pagination>
      </clr-dg-footer>
    </clr-datagrid>
  </div>
</div>
<confirmation-dialog #confirmationDialog (confirmAction)="confirmDeletion($event)"></confirmation-dialog>
</div>
`;