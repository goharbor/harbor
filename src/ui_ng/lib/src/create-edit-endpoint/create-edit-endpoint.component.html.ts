export const CREATE_EDIT_ENDPOINT_TEMPLATE: string = `
<clr-modal [(clrModalOpen)]="createEditDestinationOpened" [clrModalStaticBackdrop]="staticBackdrop" [clrModalClosable]="closable">
  <h3 class="modal-title">{{modalTitle}}</h3>
  <hbr-inline-alert class="modal-title" (confirmEvt)="confirmCancel($event)"></hbr-inline-alert>
  <div class="modal-body">
    <div class="alert alert-warning" *ngIf="!editable">
      <div class="alert-item">
        <span class="alert-text">
          {{'DESTINATION.CANNOT_EDIT' | translate}}
        </span>
      </div>
    </div>
    <form #targetForm="ngForm">
      <section class="form-block">
        <div class="form-group">
          <label for="destination_name" class="col-md-4 form-group-label-override required">{{ 'DESTINATION.NAME' | translate }}</label>
          <label class="col-md-8" for="destination_name" aria-haspopup="true" role="tooltip" [class.invalid]="targetName.errors && (targetName.dirty || targetName.touched)" [class.valid]="targetName.valid" class="tooltip tooltip-validation tooltip-sm tooltip-bottom-left">
            <input type="text" id="destination_name" [disabled]="testOngoing" [readonly]="!editable" [(ngModel)]="target.name" name="targetName" size="20" #targetName="ngModel" required  (keyup)="changedTargetName($event)"> 
            <span class="tooltip-content" *ngIf="targetName.errors && targetName.errors.required && (targetName.dirty || targetName.touched)">
              {{ 'DESTINATION.NAME_IS_REQUIRED' | translate }}
            </span>
          </label>
        </div>
        <div class="form-group">
          <label for="destination_url" class="col-md-4 form-group-label-override required">{{ 'DESTINATION.URL' | translate }}</label>
          <label class="col-md-8" for="destination_url" aria-haspopup="true" role="tooltip" [class.invalid]="targetEndpoint.errors && (targetEndpoint.dirty || targetEndpoint.touched)" [class.valid]="targetEndpoint.valid" class="tooltip tooltip-validation tooltip-sm tooltip-bottom-left">
            <input type="text" id="destination_url" [disabled]="testOngoing" [readonly]="!editable" [(ngModel)]="target.endpoint" size="20" name="endpointUrl" #targetEndpoint="ngModel" required (keyup)="clearPassword($event)" placeholder="http(s)://192.168.1.1">
            <span class="tooltip-content" *ngIf="targetEndpoint.errors && targetEndpoint.errors.required && (targetEndpoint.dirty || targetEndpoint.touched)">
              {{ 'DESTINATION.URL_IS_REQUIRED' | translate }}
            </span>
          </label>
        </div>
        <div class="form-group">
          <label for="destination_username" class="col-md-4 form-group-label-override">{{ 'DESTINATION.USERNAME' | translate }}</label>
          <input type="text" class="col-md-8" id="destination_username" [disabled]="testOngoing" [readonly]="!editable" [(ngModel)]="target.username" size="20" name="username" #username="ngModel" (keyup)="clearPassword($event)">
        </div>
        <div class="form-group">
          <label for="destination_password" class="col-md-4 form-group-label-override">{{ 'DESTINATION.PASSWORD' | translate }}</label>
          <input type="password" class="col-md-8" id="destination_password" [disabled]="testOngoing" [readonly]="!editable" [(ngModel)]="target.password" size="20" name="password" #password="ngModel" (focus)="clearPassword($event)">
        </div>
        <div class="form-group">
          <label for="destination_insecure" class="col-md-4 form-group-label-override">{{'CONFIG.VERIFY_REMOTE_CERT' | translate }}</label>
          <clr-checkbox #insecure  class="col-md-8" name="insecure" id="destination_insecure" [clrDisabled]="testOngoing" [clrChecked]="!target.insecure" (clrCheckedChange)="setInsecureValue($event)">
             <a href="javascript:void(0)" role="tooltip" aria-haspopup="true" class="tooltip tooltip-top-right" style="top:-7px;">
                    <clr-icon shape="info-circle" class="info-tips-icon" size="24"></clr-icon>
                    <span class="tooltip-content">{{'CONFIG.TOOLTIP.VERIFY_REMOTE_CERT' | translate}}</span>
                 </a>
          </clr-checkbox>
        </div>
        <div class="form-group">
          <label for="spin" class="col-md-4"></label>
          <span class="col-md-8 spinner spinner-inline" [hidden]="!inProgress"></span>
        </div>
      </section>
    </form>
  </div>
  <div class="modal-footer">
      <button type="button" class="btn btn-outline" (click)="testConnection()" [disabled]="testOngoing || onGoing || targetEndpoint.errors">{{ 'DESTINATION.TEST_CONNECTION' | translate }}</button>
      <button type="button" class="btn btn-outline" (click)="onCancel()"  [disabled]="testOngoing || onGoing">{{ 'BUTTON.CANCEL' | translate }}</button>
      <button type="submit" class="btn btn-primary" (click)="onSubmit()"  [disabled]="!isValid">{{ 'BUTTON.OK' | translate }}</button>
  </div>
</clr-modal>`;