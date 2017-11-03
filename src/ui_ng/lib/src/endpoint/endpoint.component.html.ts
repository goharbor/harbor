export const ENDPOINT_TEMPLATE: string = `
<div>
    <div class="row">
        <div class="col-lg-12 col-md-12 col-sm-12 col-xs-12">
            <div class="row flex-items-xs-between" style="height: 24px;">
                <div class="flex-items-xs-middle option-left">
                    <button class="btn btn-link" (click)="openModal()"><clr-icon shape="add"></clr-icon> {{'DESTINATION.ENDPOINT' | translate}}</button>
                    <create-edit-endpoint (reload)="reload($event)"></create-edit-endpoint>
                </div>
                <div class="flex-items-xs-middle option-right">
                    <hbr-filter [withDivider]="true" filterPlaceholder='{{"REPLICATION.FILTER_TARGETS_PLACEHOLDER" | translate}}' (filter)="doSearchTargets($event)" [currentValue]="targetName"></hbr-filter>
                    <span class="refresh-btn" (click)="refreshTargets()">
                        <clr-icon shape="refresh"></clr-icon>
                    </span>
                </div>
            </div>
        </div>
        <div class="col-lg-12 col-md-12 col-sm-12 col-xs-12">
            <clr-datagrid [clrDgLoading]="loading">
                <clr-dg-column [clrDgField]="'name'">{{'DESTINATION.NAME' | translate}}</clr-dg-column>
                <clr-dg-column [clrDgField]="'endpoint'">{{'DESTINATION.URL' | translate}}</clr-dg-column>
                <clr-dg-column [clrDgField]="'insecure'">{{'CONFIG.VERIFY_REMOTE_CERT' | translate }}</clr-dg-column>
                <clr-dg-column [clrDgSortBy]="creationTimeComparator">{{'DESTINATION.CREATION_TIME' | translate}}</clr-dg-column>
                <clr-dg-placeholder>{{'DESTINATION.PLACEHOLDER' | translate }}</clr-dg-placeholder>
                <clr-dg-row *clrDgItems="let t of targets" [clrDgItem]='t'>
                    <clr-dg-action-overflow>
                        <button class="action-item" (click)="editTarget(t)">{{'DESTINATION.TITLE_EDIT' | translate}}</button>
                        <button class="action-item" (click)="deleteTarget(t)">{{'DESTINATION.DELETE' | translate}}</button>
                    </clr-dg-action-overflow>
                    <clr-dg-cell>{{t.name}}</clr-dg-cell>
                    <clr-dg-cell>{{t.endpoint}}</clr-dg-cell>
                    <clr-dg-cell> 
                     {{!t.insecure}}
                    </clr-dg-cell>
                    <clr-dg-cell>{{t.creation_time | date: 'short'}}</clr-dg-cell>
                </clr-dg-row>
                <clr-dg-footer>
                  <span *ngIf="pagination.totalItems">{{pagination.firstItem + 1}} - {{pagination.lastItem + 1}} {{'DESTINATION.OF' | translate}}</span>
                  {{pagination.totalItems}} {{'DESTINATION.ITEMS' | translate}}
                  <clr-dg-pagination #pagination [clrDgPageSize]="15"></clr-dg-pagination>
                </clr-dg-footer>
            </clr-datagrid>
        </div>
    </div>
    <confirmation-dialog #confirmationDialog (confirmAction)="confirmDeletion($event)"></confirmation-dialog>
</div>
`;