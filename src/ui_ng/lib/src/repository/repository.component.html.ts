export const REPOSITORY_TEMPLATE = `
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
    <hbr-list-repository [urlPrefix]="urlPrefix" [projectId]="projectId" [repositories]="changedRepositories" (delete)="deleteRepo($event)" [hasProjectAdminRole]="hasProjectAdminRole" (paginate)="retrieve($event)"></hbr-list-repository>
  </div>
</div>`;