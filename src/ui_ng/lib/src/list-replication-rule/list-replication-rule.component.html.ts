export const LIST_REPLICATION_RULE_TEMPLATE: string = `
<div style="padding-bottom: 15px;">
<clr-datagrid [clrDgLoading]="loading"  [(clrDgSelected)]="selectedRow" (clrDgSelectedChange)="selectedChange()">
    <clr-dg-action-bar style="height:24px;">
        <div class="btn-group" *ngIf="opereateAvailable || isSystemAdmin">
            <button type="button" class="btn btn-sm btn-secondary" (click)="openModal()">{{'REPLICATION.NEW_REPLICATION_RULE' | translate}}</button>
            <button type="button" class="btn btn-sm btn-secondary" [disabled]="selectedRow.length !== 1" (click)="editRule(selectedRow)">{{'REPLICATION.EDIT_POLICY' | translate}}</button>
            <button type="button" class="btn btn-sm btn-secondary" [disabled]="selectedRow.length == 0" (click)="deleteRule(selectedRow)">{{'REPLICATION.DELETE_POLICY' | translate}}</button>
            <button type="button" class="btn btn-sm btn-secondary" [disabled]="selectedRow.length == 0" (click)="replicateRule(selectedRow)">{{'REPLICATION.REPLICATE' | translate}}</button>
        </div>
    </clr-dg-action-bar>
    <clr-dg-column [clrDgField]="'name'">{{'REPLICATION.NAME' | translate}}</clr-dg-column>
    <clr-dg-column [clrDgField]="'projects'" *ngIf="!projectScope">{{'REPLICATION.PROJECT' | translate}}</clr-dg-column>
    <clr-dg-column [clrDgField]="'description'">{{'REPLICATION.DESCRIPTION' | translate}}</clr-dg-column>
    <clr-dg-column [clrDgField]="'targets'">{{'REPLICATION.DESTINATION_NAME' | translate}}</clr-dg-column>
    <clr-dg-column [clrDgField]="'trigger'">{{'REPLICATION.TRIGGER_MODE' | translate}}</clr-dg-column>
    <clr-dg-placeholder>{{'REPLICATION.PLACEHOLDER' | translate }}</clr-dg-placeholder>
    <clr-dg-row *clrDgItems="let p of changedRules" [clrDgItem]="p" [style.backgroundColor]="(projectScope && withReplicationJob && selectedId === p.id) ? '#eee' : ''">
        <clr-dg-cell (click)="selectRule(p)">{{p.name}}</clr-dg-cell>
        <clr-dg-cell *ngIf="!projectScope" (click)="selectRule(p)">
            <a href="javascript:void(0)" (click)="redirectTo(p)">{{p.projects?.length>0  ? p.projects[0].name : ''}}</a>
        </clr-dg-cell>
        <clr-dg-cell (click)="selectRule(p)">{{p.description ? p.description : '-'}}</clr-dg-cell>
        <clr-dg-cell (click)="selectRule(p)">{{p.targets?.length>0 ? p.targets[0].name : ''}}</clr-dg-cell>
        <clr-dg-cell (click)="selectRule(p)">{{p.trigger ? p.trigger.kind : ''}}</clr-dg-cell>
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