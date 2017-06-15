export const COPY_INPUT_HTML: string = `
<div>
    <div class="command-title">
        {{headerTitle}}
    </div>
    <div>
        <span>
            <input type="text" class="command-input" size="{{inputSize}}" [(ngModel)]="defaultValue" #inputTarget readonly/>
        </span>
        <span>
            <clr-icon shape="copy" [class.is-success]="isCopied" [class.is-error]="hasCopyError" class="info-tips-icon" size="24" [ngxClipboard]="inputTarget" (cbOnSuccess)="onSuccess($event)" (cbOnError)="onError($event)"></clr-icon>
        </span>
    </div>
</div>
`;