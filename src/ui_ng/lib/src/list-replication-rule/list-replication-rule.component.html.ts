export const LIST_REPLICATION_RULE_TEMPLATE: string = `
<confirmation-dialog #toggleConfirmDialog (confirmAction)="toggleConfirm($event)"></confirmation-dialog>
<confirmation-dialog #deletionConfirmDialog (confirmAction)="deletionConfirm($event)"></confirmation-dialog>
<clr-datagrid>
    <clr-dg-column>{{'REPLICATION.NAME' | translate}}</clr-dg-column>
    <clr-dg-column *ngIf="projectless">{{'REPLICATION.PROJECT' | translate}}</clr-dg-column>
    <clr-dg-column>{{'REPLICATION.DESCRIPTION' | translate}}</clr-dg-column>
    <clr-dg-column>{{'REPLICATION.DESTINATION_NAME' | translate}}</clr-dg-column>
    <clr-dg-column>{{'REPLICATION.LAST_START_TIME' | translate}}</clr-dg-column>
    <clr-dg-column>{{'REPLICATION.ACTIVATION' | translate}}</clr-dg-column>
    <clr-dg-row *clrDgItems="let p of rules" [clrDgItem]="p" (click)="selectRule(p)" [style.backgroundColor]="(!projectless && selectedId === p.id) ? '#eee' : ''">
        <clr-dg-action-overflow>
            <button class="action-item" (click)="editRule(p)">{{'REPLICATION.EDIT_POLICY' | translate}}</button>
            <button class="action-item" (click)="toggleRule(p)">{{ (p.enabled === 0 ? 'REPLICATION.ENABLE' : 'REPLICATION.DISABLE') | translate}}</button>
            <button class="action-item" (click)="deleteRule(p)">{{'REPLICATION.DELETE_POLICY' | translate}}</button>
        </clr-dg-action-overflow>
        <clr-dg-cell>
          <ng-template [ngIf]="projectless">
            <a href="javascript:void(0)">{{p.name}}</a>
          </ng-template>
          <ng-template [ngIf]="!projectless">
            {{p.name}}
          </ng-template>
        </clr-dg-cell>
        <clr-dg-cell *ngIf="projectless">{{p.project_name}}</clr-dg-cell>
        <clr-dg-cell>{{p.description ? p.description : '-'}}</clr-dg-cell>
        <clr-dg-cell>{{p.target_name}}</clr-dg-cell>
        <clr-dg-cell>
          <ng-template [ngIf]="p.start_time === nullTime">-</ng-template>
          <ng-template [ngIf]="p.start_time !== nullTime">{{p.start_time}}</ng-template>
        </clr-dg-cell>
        <clr-dg-cell>
            {{ (p.enabled === 1 ? 'REPLICATION.ENABLED' : 'REPLICATION.DISABLED') | translate}}
        </clr-dg-cell>
    </clr-dg-row>
    <clr-dg-footer>{{ (rules ? rules.length : 0) }} {{'REPLICATION.ITEMS' | translate}}</clr-dg-footer>
</clr-datagrid>`;