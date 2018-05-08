import { Component, Input, Output, EventEmitter, OnInit } from '@angular/core';


export const enum CopyStatus {
    NORMAL, SUCCESS, ERROR
}

@Component({
    selector: 'hbr-copy-input',
    templateUrl: './copy-input.coponent.html',
    styleUrls: ['./push-image.scss'],

    providers: []
})

export class CopyInputComponent implements OnInit {
    @Input() inputSize: number = 40;
    @Input() headerTitle: string = "Copy Input";
    @Input() defaultValue: string = "N/A";
    @Input() iconMode: boolean = false;

    state: CopyStatus = CopyStatus.NORMAL;

    @Output() onCopySuccess: EventEmitter<any> = new EventEmitter<any>();
    @Output() onCopyError: EventEmitter<any> = new EventEmitter<any>();

    ngOnInit(): void { }
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
