export const REGISTRY_CONFIG_HTML: string = `
<div>
    <system-settings #systemSettings [(systemSettings)]="config" [showSubTitle]="true" [hasAdminRole]="hasAdminRole" [hasCAFile]="hasCAFile"></system-settings>
    <vulnerability-config *ngIf="withClair" #vulnerabilityConfig [(vulnerabilityConfig)]="config" [showSubTitle]="true"></vulnerability-config>
    <div>
        <button type="button" class="btn btn-primary" (click)="save()" [disabled]="shouldDisable">{{'BUTTON.SAVE' | translate}}</button>
        <button type="button" class="btn btn-outline" (click)="cancel()" [disabled]="shouldDisable">{{'BUTTON.CANCEL' | translate}}</button>
    </div>
    <confirmation-dialog #cfgConfirmationDialog (confirmAction)="confirmCancel($event)"></confirmation-dialog>
</div>
`;