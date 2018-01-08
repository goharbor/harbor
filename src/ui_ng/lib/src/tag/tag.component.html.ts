export const TAG_TEMPLATE = `
<confirmation-dialog class="hidden-tag" #confirmationDialog  [batchInfors]="batchDelectionInfos" (confirmAction)="confirmDeletion($event)"></confirmation-dialog>
<clr-modal class="hidden-tag" [(clrModalOpen)]="showTagManifestOpened" [clrModalStaticBackdrop]="staticBackdrop" [clrModalClosable]="closable">
  <h3 class="modal-title">{{ manifestInfoTitle | translate }}</h3>
  <div class="modal-body">
    <div class="row col-md-12">
        <textarea rows="2" #digestTarget>{{digestId}}</textarea>
    </div>
  </div>
  <div class="modal-footer">
    <span class="copy-failed" [hidden]="!copyFailed">{{'TAG.COPY_ERROR' | translate}}</span>
    <button type="button" class="btn btn-primary" [ngxClipboard]="digestTarget" (cbOnSuccess)="onSuccess($event)" (cbOnError)="onError($event)">{{'BUTTON.COPY' | translate}}</button>
  </div>
</clr-modal>
<div class="row" style="position:relative;">
  <div>
    <div class="row flex-items-xs-right rightPos">
      <div class="flex-xs-middle">
        <hbr-filter [withDivider]="true" filterPlaceholder="{{'TAG.FILTER_FOR_TAGS' | translate}}" (filter)="doSearchTagNames($event)" [currentValue]="lastFilteredTagName"></hbr-filter>
        <span class="refresh-btn" (click)="refresh()"><clr-icon shape="refresh"></clr-icon></span> 
      </div>
    </div>
  </div>
  <div class="col-lg-12 col-md-12 col-sm-12 col-xs-12"> 
    <clr-datagrid [clrDgLoading]="loading" [class.embeded-datagrid]="isEmbedded"  [(clrDgSelected)]="selectedRow" (clrDgSelectedChange)="selectedChange()">
        <clr-dg-action-bar>
            <div class="btn-group">
                <button type="button" class="btn btn-sm btn-secondary" [disabled]="!canScanNow(selectedRow)" [disabled]="!(selectedRow.length==1)" (click)="scanNow(selectedRow)">{{'VULNERABILITY.SCAN_NOW' | translate}}</button>
                <button type="button" class="btn btn-sm btn-secondary" [disabled]="!(selectedRow.length==1)" (click)="showDigestId(selectedRow)" >{{'REPOSITORY.COPY_DIGEST_ID' | translate}}</button>
                <button type="button" class="btn btn-sm btn-secondary" *ngIf="hasProjectAdminRole" (click)="deleteTags(selectedRow)" [disabled]="!selectedRow.length">{{'REPOSITORY.DELETE' | translate}}</button>
            </div>
        </clr-dg-action-bar>
        <clr-dg-column style="width: 160px;" [clrDgField]="'name'">{{'REPOSITORY.TAG' | translate}}</clr-dg-column>
        <clr-dg-column style="width: 90px;" [clrDgField]="'size'">{{'REPOSITORY.SIZE' | translate}}</clr-dg-column>
        <clr-dg-column style="min-width: 120px; max-width:220px;">{{'REPOSITORY.PULL_COMMAND' | translate}}</clr-dg-column>
        <clr-dg-column style="width: 140px;" *ngIf="withClair">{{'REPOSITORY.VULNERABILITY' | translate}}</clr-dg-column>
        <clr-dg-column style="width: 80px;" *ngIf="withNotary">{{'REPOSITORY.SIGNED' | translate}}</clr-dg-column>
        <clr-dg-column style="min-width: 130px;">{{'REPOSITORY.AUTHOR' | translate}}</clr-dg-column>
        <clr-dg-column style="width: 160px;"[clrDgSortBy]="createdComparator">{{'REPOSITORY.CREATED' | translate}}</clr-dg-column>
        <clr-dg-column style="width: 80px;" [clrDgField]="'docker_version'" *ngIf="!withClair">{{'REPOSITORY.DOCKER_VERSION' | translate}}</clr-dg-column>
        <clr-dg-placeholder>{{'TGA.PLACEHOLDER' | translate }}</clr-dg-placeholder>
        <clr-dg-row *ngFor="let t of tags" [clrDgItem]='t'>
          <clr-dg-cell  class="truncated"  style="width: 160px;" [ngSwitch]="withClair">
            <a *ngSwitchCase="true" href="javascript:void(0)" (click)="onTagClick(t)" title="{{t.name}}">{{t.name}}</a>
            <span *ngSwitchDefault>{{t.name}}</span>
          </clr-dg-cell>
          <clr-dg-cell style="width: 90px;">{{t.size}}</clr-dg-cell>
          <clr-dg-cell style="min-width: 120px; max-width:220px;" class="truncated" title="docker pull {{registryUrl}}/{{repoName}}:{{t.name}}">
              <hbr-copy-input #copyInput  (onCopyError)="onCpError($event)"  iconMode="true" defaultValue="docker pull {{registryUrl}}/{{repoName}}:{{t.name}}"></hbr-copy-input>
          </clr-dg-cell>
          <clr-dg-cell style="width: 140px;" *ngIf="withClair">
            <hbr-vulnerability-bar [repoName]="repoName" [tagId]="t.name" [summary]="t.scan_overview"></hbr-vulnerability-bar>
          </clr-dg-cell>
          <clr-dg-cell style="width: 80px;" *ngIf="withNotary"  [ngSwitch]="t.signature !== null">
            <clr-icon shape="check-circle" *ngSwitchCase="true"  size="20" style="color: #1D5100;"></clr-icon>
            <clr-icon shape="times-circle" *ngSwitchCase="false"  size="16" style="color: #C92100;"></clr-icon>
            <a href="javascript:void(0)" *ngSwitchDefault role="tooltip" aria-haspopup="true" class="tooltip tooltip-top-right">
              <clr-icon shape="help" style="color: #565656;" size="16"></clr-icon>
              <span class="tooltip-content">{{'REPOSITORY.NOTARY_IS_UNDETERMINED' | translate}}</span>
            </a>
          </clr-dg-cell>
          <clr-dg-cell  class="truncated"  style="min-width: 130px;" title="{{t.author}}">{{t.author}}</clr-dg-cell>
          <clr-dg-cell style="width: 160px;">{{t.created | date: 'short'}}</clr-dg-cell>
          <clr-dg-cell style="width: 80px;" *ngIf="!withClair">{{t.docker_version}}</clr-dg-cell>
        </clr-dg-row>
        <clr-dg-footer> 
          <span *ngIf="pagination.totalItems">{{pagination.firstItem + 1}} - {{pagination.lastItem + 1}} {{'REPOSITORY.OF' | translate}}</span>
          {{pagination.totalItems}} {{'REPOSITORY.ITEMS' | translate}}&nbsp;&nbsp;&nbsp;&nbsp;
          <clr-dg-pagination #pagination [clrDgPageSize]="10"></clr-dg-pagination>
        </clr-dg-footer>
    </clr-datagrid>
  </div>
</div>`;
