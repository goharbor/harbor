export const REGISTRY_CONFIG_HTML: string = `
<div>
    <replication-config #replicationConfig [(replicationConfig)]="config"></replication-config>
    <system-settings #systemSettings [(systemSettings)]="config"></system-settings>
    <vulnerability-config #vulnerabilityConfig [(vulnerabilityConfig)]="config"></vulnerability-config>
    <div>
        <button type="button" class="btn btn-primary" (click)="save()" [disabled]="shouldDisable">{{'BUTTON.SAVE' | translate}}</button>
        <button type="button" class="btn btn-outline" (click)="cancel()" [disabled]="shouldDisable">{{'BUTTON.CANCEL' | translate}}</button>
    </div>
    <confirmation-dialog #cfgConfirmationDialog (confirmAction)="confirmCancel($event)"></confirmation-dialog>
</div>
`;