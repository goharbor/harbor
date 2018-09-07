import { ClipboardService } from './clipboard.service';
import { Directive, EventEmitter, HostListener, Input, OnDestroy, OnInit, Output, Renderer, ElementRef } from '@angular/core';

@Directive({
    selector: '[ngxClipboard]'
})
export class ClipboardDirective implements OnInit, OnDestroy {
    // tslint:disable-next-line:no-input-rename
    @Input('ngxClipboard') public targetElm: HTMLInputElement;

    @Input() public cbContent: string;

    @Output() public cbOnSuccess: EventEmitter<any> = new EventEmitter<any>();

    @Output() public cbOnError: EventEmitter<any> = new EventEmitter<any>();
    constructor(
        private clipboardSrv: ClipboardService,
        private renderer: Renderer

    ) { }

    public ngOnInit() { }

    public ngOnDestroy() {
        this.clipboardSrv.destroy();
    }

    @HostListener('click', ['$event.target'])
    // tslint:disable-next-line:no-unused-variable
    public onClick(button: ElementRef) {
        if (!this.clipboardSrv.isSupported) {
            this.handleResult(false, undefined);
        } else if (this.targetElm && this.clipboardSrv.isTargetValid(this.targetElm)) {
            this.handleResult(this.clipboardSrv.copyFromInputElement(this.targetElm, this.renderer),
                this.targetElm.value);
        } else if (this.cbContent) {
            this.handleResult(this.clipboardSrv.copyFromContent(this.cbContent, this.renderer), this.cbContent);
        }
    }

    /**
     * Fires an event based on the copy operation result.
     *  ** deprecated param {Boolean} succeeded
     */
    private handleResult(succeeded: Boolean, copiedContent: string) {
        if (succeeded) {
            this.cbOnSuccess.emit({ isSuccess: true, content: copiedContent });
        } else {
            this.cbOnError.emit({ isSuccess: false });
        }
    }
}
