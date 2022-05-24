import { Component, Input, Output, EventEmitter } from '@angular/core';

export const enum CopyStatus {
    NORMAL,
    SUCCESS,
    ERROR,
}

@Component({
    selector: 'hbr-copy-input',
    templateUrl: './copy-input.coponent.html',
    styleUrls: ['./push-image.scss'],

    providers: [],
})
export class CopyInputComponent {
    @Input() inputSize: number = 40;
    @Input() headerTitle: string = 'Copy Input';
    @Input() defaultValue: string = 'N/A';
    @Input() iconMode: boolean = false;

    state: CopyStatus = CopyStatus.NORMAL;

    // eslint-disable-next-line @angular-eslint/no-output-on-prefix
    @Output() onCopySuccess: EventEmitter<any> = new EventEmitter<any>();
    // eslint-disable-next-line @angular-eslint/no-output-on-prefix
    @Output() onCopyError: EventEmitter<any> = new EventEmitter<any>();
    onSuccess($event: any): void {
        this.state = CopyStatus.SUCCESS;
        this.onCopySuccess.emit($event);
    }

    onError(error: any): void {
        this.state = CopyStatus.ERROR;
        this.onCopyError.emit(error);
    }

    reset(): void {
        this.state = CopyStatus.NORMAL;
    }

    setPullCommendShow(): void {
        this.iconMode = false;
    }

    public get isCopied(): boolean {
        return this.state === CopyStatus.SUCCESS;
    }

    public get hasCopyError(): boolean {
        return this.state === CopyStatus.ERROR;
    }
}
