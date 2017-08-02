export const JOB_LOG_VIEWER_TEMPLATE: string = `
<clr-modal [(clrModalOpen)]="opened" [clrModalStaticBackdrop]="true" [clrModalSize]="'xl'">
    <h3 class="modal-title" class="log-viewer-title" style="margin-top: 0px;">{{title | translate }}</h3>
    <div class="modal-body">
      <div class="loading-back" [hidden]="!onGoing">
        <span class="spinner spinner-md"></span>
      </div>
      <pre [hidden]="onGoing">
<code>{{log}}</code>
      </pre>
    </div>
    <div class="modal-footer">
      <button type="button" class="btn btn-primary" (click)="close()">{{ 'BUTTON.CLOSE' | translate}}</button>
    </div>
</clr-modal>
`;

export const JOB_LOG_VIEWER_STYLES: string = `
.log-viewer-title {
    line-height: 24px;
    color: #000000;
    font-size: 22px;
}

.loading-back {
  height: 358px;
  display: flex;
  align-items: center;
  justify-content: center; 
}
`;