export const REPOSITORY_STACKVIEW_TEMPLATE: string = `
<confirmation-dialog #confirmationDialog (confirmAction)="confirmDeletion($event)"></confirmation-dialog>
<div class="row">
  <div class="col-lg-12 col-md-12 col-sm-12 col-xs-12">  
    <div class="row flex-items-xs-right option-right">
      <div class="flex-xs-middle">
        <hbr-filter filterPlaceholder="{{'REPOSITORY.FILTER_FOR_REPOSITORIES' | translate}}" (filter)="doSearchRepoNames($event)"></hbr-filter>  
        <a href="javascript:void(0)" (click)="refresh()"><clr-icon shape="refresh"></clr-icon></a>
      </div>
    </div>
  </div>
  <div class="col-lg-12 col-md-12 col-sm-12 col-xs-12">  
    <clr-datagrid>
      <clr-dg-column [clrDgField]="'name'">{{'REPOSITORY.NAME' | translate}}</clr-dg-column>
      <clr-dg-column [clrDgSortBy]="tagsCountComparator">{{'REPOSITORY.TAGS_COUNT' | translate}}</clr-dg-column>
      <clr-dg-column [clrDgSortBy]="pullCountComparator">{{'REPOSITORY.PULL_COUNT' | translate}}</clr-dg-column>
      <clr-dg-row *clrDgItems="let r of repositories">
        <clr-dg-action-overflow [hidden]="!hasProjectAdminRole">
          <button class="action-item" (click)="deleteRepo(r.name)">{{'REPOSITORY.DELETE' | translate}}</button>
        </clr-dg-action-overflow>
        <clr-dg-cell>{{r.name}}</clr-dg-cell>
        <clr-dg-cell>{{r.tags_count}}</clr-dg-cell>
        <clr-dg-cell>{{r.pull_count}}</clr-dg-cell>        
        <hbr-tag *clrIfExpanded ngProjectAs="clr-dg-row-detail" class="sub-grid-custom" [repoName]="r.name" [sessionInfo]="sessionInfo" [projectId]="projectId" [isEmbeded]="true" (refreshRepo)="refresh($event)"></hbr-tag>
      </clr-dg-row>
      <clr-dg-footer> 
        {{ repositories ? repositories.length : 0 }} {{'REPOSITORY.ITEMS' | translate}}
      </clr-dg-footer>
    </clr-datagrid>
  </div>
</div>
`;