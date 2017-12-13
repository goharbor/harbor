export const CREATE_EDIT_RULE_TEMPLATE: string = `
<clr-modal [(clrModalOpen)]="createEditRuleOpened" [clrModalStaticBackdrop]="staticBackdrop" [clrModalClosable]="closable">
  <h3 class="modal-title">{{modalTitle}}</h3>
  <hbr-inline-alert class="modal-title" (confirmEvt)="confirmCancel($event)"></hbr-inline-alert>
  <div class="modal-body" style="max-height: 85vh;">
    <form #ruleForm="ngForm">
      <section class="form-block">
        <div class="alert alert-warning" *ngIf="!editable">
          <div class="alert-item">
            <span class="alert-text">
              {{'REPLICATION.CANNOT_EDIT' | translate}}
            </span>
          </div>
        </div>
        <div class="form-group">
          <label for="policy_name" class="col-md-4 form-group-label-override">{{'REPLICATION.NAME' | translate}}<span style="color: red">*</span></label>
          <label for="policy_name" class="col-md-8"  aria-haspopup="true" role="tooltip" [class.invalid]="name.errors && (name.dirty || name.touched)" class="tooltip tooltip-validation tooltip-sm tooltip-bottom-left">
            <input type="text" id="policy_name" [(ngModel)]="createEditRule.name" name="name" size="20" #name="ngModel" required [readonly]="readonly">
            <span class="tooltip-content" *ngIf="name.errors && name.errors.required && (name.dirty || name.touched)">
              {{'REPLICATION.NAME_IS_REQUIRED' | translate}}
            </span>
          </label>
        </div>
        <div class="form-group">
          <label for="policy_description" class="col-md-4 form-group-label-override">{{'REPLICATION.DESCRIPTION' | translate}}</label>
          <textarea class="col-md-8" id="policy_description" row="3" [(ngModel)]="createEditRule.description" name="description" size="20" #description="ngModel" [readonly]="readonly"></textarea>
        </div>
        <div class="form-group">
          <label class="col-md-4 form-group-label-override">{{'REPLICATION.ENABLE' | translate}}</label>
          <div class="checkbox-inline">
            <input type="checkbox" id="policy_enable" [(ngModel)]="createEditRule.enable" name="enable" #enable="ngModel" [disabled]="untoggleable">
            <label for="policy_enable"></label>
          </div>
        </div>
        <div class="form-group">
          <label for="destination_name" class="col-md-4 form-group-label-override">{{'REPLICATION.DESTINATION_NAME' | translate}}<span style="color: red">*</span></label>
          <div class="select" *ngIf="!isCreateEndpoint">
            <select id="destination_name" [(ngModel)]="createEditRule.endpointId" name="endpointId" (change)="selectEndpoint()" [disabled]="testOngoing || readonly">
              <option *ngFor="let t of endpoints" [value]="t.id" [selected]="t.id == createEditRule.endpointId">{{t.name}}</option>
            </select>
          </div>
          <label class="col-md-8" *ngIf="isCreateEndpoint" for="destination_name" aria-haspopup="true" role="tooltip" [class.invalid]="endpointName.errors && (endpointName.dirty || endpointName.touched)"
            class="tooltip tooltip-validation tooltip-sm tooltip-bottom-left">
            <input type="text" id="destination_name" [(ngModel)]="createEditRule.endpointName" name="endpointName" size="8" #endpointName="ngModel" value="" required> 
            <span class="tooltip-content" *ngIf="endpointName.errors && endpointName.errors.required && (endpointName.dirty || endpointName.touched)">
              {{'REPLICATION.DESTINATION_NAME_IS_REQUIRED' | translate}}
            </span>
          </label>
          <div class="checkbox-inline" *ngIf="showNewDestination">
            <input type="checkbox" id="check_new" (click)="newEndpoint(checkedAddNew.checked)" #checkedAddNew [checked]="isCreateEndpoint" [disabled]="testOngoing || readonly">
            <label for="check_new">{{'REPLICATION.NEW_DESTINATION' | translate}}</label>
          </div>
        </div>
        <div class="form-group">
          <label for="destination_url" class="col-md-4 form-group-label-override">{{'REPLICATION.DESTINATION_URL' | translate}}<span style="color: red">*</span></label>
          <label for="destination_url" class="col-md-8" aria-haspopup="true" role="tooltip" [class.invalid]="endpointUrl.errors && (endpointUrl.dirty || endpointUrl.touched)"
            class="tooltip tooltip-validation tooltip-sm tooltip-bottom-left">
            <input type="text" id="destination_url" [disabled]="testOngoing" [readonly]="readonly || !isCreateEndpoint"
            [(ngModel)]="createEditRule.endpointUrl" size="20" name="endpointUrl" required #endpointUrl="ngModel">
            <span class="tooltip-content" *ngIf="endpointUrl.errors && endpointUrl.errors.required && (endpointUrl.dirty || endpointUrl.touched)">
              {{'REPLICATION.DESTINATION_URL_IS_REQUIRED' | translate}}
            </span>
          </label>
        </div>
        <div class="form-group">
          <label for="destination_username" class="col-md-4 form-group-label-override">{{'REPLICATION.DESTINATION_USERNAME' | translate}}</label>
          <input type="text" class="col-md-8" id="destination_username" [disabled]="testOngoing" [readonly]="readonly || !isCreateEndpoint" 
          [(ngModel)]="createEditRule.username" size="20" name="username" #username="ngModel">
        </div>
        <div class="form-group">
          <label for="destination_password" class="col-md-4 form-group-label-override">{{'REPLICATION.DESTINATION_PASSWORD' | translate}}</label>
          <input type="password" class="col-md-8" id="destination_password" [disabled]="testOngoing" [readonly]="readonly || !isCreateEndpoint" 
          [(ngModel)]="createEditRule.password" size="20" name="password" #password="ngModel">
        </div>
        <div class="form-group">
          <label for="destination_insecure" class="col-md-4 form-group-label-override">{{'CONFIG.VERIFY_REMOTE_CERT' | translate }}</label>
          <clr-checkbox #insecure  class="col-md-8" name="insecure" id="destination_insecure" [clrChecked]="!createEditRule.insecure"  [clrDisabled]="readonly || !isCreateEndpoint || testOngoing" (clrCheckedChange)="setInsecureValue($event)">
             <a href="javascript:void(0)" role="tooltip" aria-haspopup="true" class="tooltip tooltip-top-right" style="top:-7px;">
                    <clr-icon shape="info-circle" class="info-tips-icon" size="24"></clr-icon>
                    <span class="tooltip-content">{{'CONFIG.TOOLTIP.VERIFY_REMOTE_CERT' | translate}}</span>
                 </a>
          </clr-checkbox>
        </div>
        <div class="form-group">
          <label for="spin" class="col-md-4"></label>
          <span class="col-md-8 spinner spinner-inline" [hidden]="!testOngoing"></span>
          <span [style.color]="!pingStatus ? 'red': ''" class="form-group-label-override">{{ pingTestMessage }}</span>
        </div>
      </section>
    </form>
  </div>
  <div class="modal-footer">
      <button type="button" class="btn btn-outline" (click)="testConnection()" [disabled]="testOngoing || endpointUrl.errors || connectAbled">{{'REPLICATION.TEST_CONNECTION' | translate}}</button>
      <button type="button" class="btn btn-outline" [disabled]="btnAbled" (click)="onCancel()">{{'BUTTON.CANCEL' | translate }}</button>
      <button type="submit" class="btn btn-primary" [disabled]="!ruleForm.form.valid || testOngoing || !editable" (click)="onSubmit()">{{'BUTTON.OK' | translate}}</button>
  </div>
</clr-modal>`;
