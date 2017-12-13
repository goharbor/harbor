export const PROJECT_POLICY_CONFIG_TEMPLATE = `
<form #projectPolicyForm="ngForm">
    <section class="form-block">
        <div class="form-group">
          <label for="projectPolicyForm">{{ 'PROJECT_CONFIG.REGISTRY' | translate }}</label>
          <div class="form-content">
            <clr-checkbox [(ngModel)]="projectPolicy.Public" name="public"
              [clrDisabled]="!hasProjectAdminRole">{{ 'PROJECT_CONFIG.PUBLIC_TOGGLE' | translate }}</clr-checkbox>
            <div>
              <label> {{ 'PROJECT_CONFIG.PUBLIC_POLICY' | translate }} </label>
            </div>
          </div>
        </div>
        <div class="form-group" *ngIf="withNotary || withClair">
          <label for="projectPolicyForm">{{ 'PROJECT_CONFIG.SECURITY' | translate }}</label>
          <div class="form-content">
            <div *ngIf="withNotary">
              <clr-checkbox [(ngModel)]="projectPolicy.ContentTrust" name="content-trust" 
                [clrDisabled]="!hasProjectAdminRole">{{ 'PROJECT_CONFIG.CONTENT_TRUST_TOGGLE' | translate }}</clr-checkbox>
              <div class="chk-explain"><label> {{ 'PROJECT_CONFIG.CONTENT_TRUST_POLCIY' | translate }} </label></div>
            </div>
            <div *ngIf="withClair">
              <clr-checkbox [(ngModel)]="projectPolicy.PreventVulImg" name="prevent-vulenrability-image" [clrDisabled]="!hasProjectAdminRole">{{ 'PROJECT_CONFIG.PREVENT_VULNERABLE_TOGGLE' | translate }}</clr-checkbox>
              <div class="chk-explain">
                <label>
                  <div id="severity-blk">
                    <div>{{ 'PROJECT_CONFIG.PREVENT_VULNERABLE_1' | translate }}</div>
                    <div class="select">
                      <select id="severity" name="severity" [(ngModel)]="projectPolicy.PreventVulImgSeverity" [disabled]="!projectPolicy.PreventVulImg">
                        <option *ngFor='let s of severityOptions' [ngValue]="s.severity">{{ s.severityLevel | translate | uppercase }}</option>                      
                      </select>
                    </div> 
                    <div> {{ 'PROJECT_CONFIG.PREVENT_VULNERABLE_2' | translate }} </div>
                  </div>
                </label>
              </div>
            </div>
          </div>
        </div>
        <div class="form-group" *ngIf="withClair">
          <label for="projectPolicyForm">{{ 'PROJECT_CONFIG.SCAN' | translate }}</label>
          <div class="form-content">
            <clr-checkbox [(ngModel)]="projectPolicy.ScanImgOnPush" name="scan-image-on-push" [clrDisabled]="!hasProjectAdminRole">{{ 'PROJECT_CONFIG.AUTOSCAN_TOGGLE' | translate }}</clr-checkbox>
            <div class="chk-explain"><label> {{ 'PROJECT_CONFIG.AUTOSCAN_POLICY' | translate }}</label></div>
          </div>
        </div>
       <button type="button" class="btn btn-primary" (click)="save()" [disabled]="!isValid() || !hasChanges() || !hasProjectAdminRole">{{'BUTTON.SAVE' | translate}}</button>
       <button type="button" class="btn btn-outline" (click)="cancel()" [disabled]="!isValid() || !hasChanges() || !hasProjectAdminRole">{{'BUTTON.CANCEL' | translate}}</button>
       <confirmation-dialog #cfgConfirmationDialog (confirmAction)="confirmCancel($event)"></confirmation-dialog>       
    </section>
</form>`;
