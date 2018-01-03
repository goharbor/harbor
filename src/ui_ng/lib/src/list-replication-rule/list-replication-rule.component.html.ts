export const LIST_REPLICATION_RULE_TEMPLATE: string = `
<div>
<clr-datagrid [clrDgLoading]="loading"  [(clrDgSingleSelected)]="selectedRow" (clrDgSingleSelectedChange)="selectedChange()">
    <clr-dg-action-bar>
        <div class="btn-group">
            <button type="button" *ngIf="creationAvailable" class="btn btn-sm btn-secondary" (click)="openModal()">{{'REPLICATION.NEW_REPLICATION_RULE' | translate}}</button>
            <button type="button" *ngIf="creationAvailable" class="btn btn-sm btn-secondary" [disabled]="!selectedRow" (click)="editRule(selectedRow)">{{'REPLICATION.EDIT_POLICY' | translate}}</button>
            <button type="button" *ngIf="creationAvailable" class="btn btn-sm btn-secondary" [disabled]="!selectedRow" (click)="deleteRule(selectedRow)">{{'REPLICATION.DELETE_POLICY' | translate}}</button>
            <button type="button" *ngIf="creationAvailable" class="btn btn-sm btn-secondary" [disabled]="!selectedRow" (click)="replicateRule(selectedRow)">{{'REPLICATION.REPLICATE' | translate}}</button>
        </div>
    </clr-dg-action-bar>
    <clr-dg-column [clrDgField]="'name'">{{'REPLICATION.NAME' | translate}}</clr-dg-column>
    <clr-dg-column [clrDgField]="'projects'" *ngIf="!projectScope">{{'REPLICATION.PROJECT' | translate}}</clr-dg-column>
    <clr-dg-column [clrDgField]="'description'">{{'REPLICATION.DESCRIPTION' | translate}}</clr-dg-column>
    <clr-dg-column [clrDgField]="'targets'">{{'REPLICATION.DESTINATION_NAME' | translate}}</clr-dg-column>
    <clr-dg-column [clrDgField]="'trigger'">{{'REPLICATION.SCHEDULE' | translate}}</clr-dg-column>
    <clr-dg-placeholder>{{'REPLICATION.PLACEHOLDER' | translate }}</clr-dg-placeholder>
    <clr-dg-row *clrDgItems="let p of changedRules" [clrDgItem]="p" (click)="selectRule(p)" [style.backgroundColor]="(projectScope && withReplicationJob && selectedId === p.id) ? '#eee' : ''">
        <clr-dg-cell>{{p.name}}</clr-dg-cell>
        <clr-dg-cell *ngIf="!projectScope">
            <a href="javascript:void(0)" (click)="redirectTo(p)">{{p.projects?.length>0  ? p.projects[0].name : ''}}</a>
        </clr-dg-cell>
        <clr-dg-cell>{{p.description ? p.description : '-'}}</clr-dg-cell>
        <clr-dg-cell>{{p.targets?.length>0 ? p.targets[0].name : ''}}</clr-dg-cell>
        <clr-dg-cell>{{p.trigger ? p.trigger.kind : ''}}</clr-dg-cell>
    </clr-dg-row>
    <clr-dg-footer>
      <span *ngIf="pagination.totalItems">{{pagination.firstItem + 1}} - {{pagination.lastItem +1 }} {{'REPLICATION.OF' | translate}} </span>{{pagination.totalItems }} {{'REPLICATION.ITEMS' | translate}}
      <clr-dg-pagination #pagination [clrDgPageSize]="5"></clr-dg-pagination>
    </clr-dg-footer>
</clr-datagrid>
<confirmation-dialog #toggleConfirmDialog  [batchInfors]="batchDelectionInfos"  (confirmAction)="toggleConfirm($event)"></confirmation-dialog>
<confirmation-dialog #deletionConfirmDialog [batchInfors]="batchDelectionInfos" (confirmAction)="deletionConfirm($event)"></confirmation-dialog>
</div>
`;