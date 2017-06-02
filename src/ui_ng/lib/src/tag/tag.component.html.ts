export const TAG_TEMPLATE = `
<confirmation-dialog class="hidden-tag" #confirmationDialog (confirmAction)="confirmDeletion($event)"></confirmation-dialog>
<clr-modal class="hidden-tag" [(clrModalOpen)]="showTagManifestOpened" [clrModalStaticBackdrop]="staticBackdrop" [clrModalClosable]="closable">
  <h3 class="modal-title">{{ manifestInfoTitle | translate }}</h3>
  <div class="modal-body">
    <div class="row col-md-12">
        <textarea rows="3" (click)="selectAndCopy($event)">{{digestId}}</textarea>
    </div>
  </div>
  <div class="modal-footer">
    <button type="button" class="btn btn-primary" (click)="showTagManifestOpened = false">{{'BUTTON.OK' | translate}}</button>
  </div>
</clr-modal>

<h2 *ngIf="!isEmbedded" class="sub-header-title">{{repoName}}</h2>
<clr-datagrid [clrDgLoading]="loading" [class.embeded-datagrid]="isEmbedded">
    <clr-dg-column [clrDgField]="'name'">{{'REPOSITORY.TAG' | translate}}</clr-dg-column>
    <clr-dg-column>{{'REPOSITORY.PULL_COMMAND' | translate}}</clr-dg-column>
    <clr-dg-column *ngIf="withNotary">{{'REPOSITORY.SIGNED' | translate}}</clr-dg-column>
    <clr-dg-column>{{'REPOSITORY.AUTHOR' | translate}}</clr-dg-column>
    <clr-dg-column [clrDgSortBy]="createdComparator">{{'REPOSITORY.CREATED' | translate}}</clr-dg-column>
    <clr-dg-column [clrDgField]="'docker_version'">{{'REPOSITORY.DOCKER_VERSION' | translate}}</clr-dg-column>
    <clr-dg-column [clrDgField]="'architecture'">{{'REPOSITORY.ARCHITECTURE' | translate}}</clr-dg-column>
    <clr-dg-column [clrDgField]="'os'">{{'REPOSITORY.OS' | translate}}</clr-dg-column>
    <clr-dg-row *clrDgItems="let t of tags" [clrDgItem]='t'>
      <clr-dg-action-overflow>
        <button class="action-item" (click)="showDigestId(t)">{{'REPOSITORY.COPY_DIGEST_ID' | translate}}</button>
        <button class="action-item" [hidden]="!hasProjectAdminRole" (click)="deleteTag(t)">{{'REPOSITORY.DELETE' | translate}}</button>
      </clr-dg-action-overflow>
      <clr-dg-cell>{{t.name}}</clr-dg-cell>
      <clr-dg-cell>docker pull {{registryUrl}}/{{repoName}}:{{t.name}}</clr-dg-cell>
      <clr-dg-cell *ngIf="withNotary"  [ngSwitch]="t.signature !== null">
        <clr-icon shape="check" *ngSwitchCase="true" style="color: #1D5100;"></clr-icon>
        <clr-icon shape="close" *ngSwitchCase="false" style="color: #C92100;"></clr-icon>
        <a href="javascript:void(0)" *ngSwitchDefault role="tooltip" aria-haspopup="true" class="tooltip tooltip-top-right">
          <clr-icon shape="help" style="color: #565656;" size="16"></clr-icon>
          <span class="tooltip-content">{{'REPOSITORY.NOTARY_IS_UNDETERMINED' | translate}}</span>
        </a>
      </clr-dg-cell>
      <clr-dg-cell>{{t.author}}</clr-dg-cell>
      <clr-dg-cell>{{t.created | date: 'short'}}</clr-dg-cell>
      <clr-dg-cell>{{t.docker_version}}</clr-dg-cell>
      <clr-dg-cell>{{t.architecture}}</clr-dg-cell>
      <clr-dg-cell>{{t.os}}</clr-dg-cell>
    </clr-dg-row>
    <clr-dg-footer> 
      {{pagination.firstItem + 1}} - {{pagination.lastItem + 1}} {{'REPOSITORY.OF' | translate}}
      {{pagination.totalItems}} {{'REPOSITORY.ITEMS' | translate}}&nbsp;&nbsp;&nbsp;&nbsp;
      <clr-dg-pagination #pagination [clrDgPageSize]="5"></clr-dg-pagination>
    </clr-dg-footer>
</clr-datagrid>`;