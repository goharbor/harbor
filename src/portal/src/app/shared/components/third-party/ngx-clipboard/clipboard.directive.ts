import { ClipboardService } from './clipboard.service';
import {
    Directive,
    EventEmitter,
    HostListener,
    Input,
    OnDestroy,
    Output,
    ElementRef,
    Renderer2,
} from '@angular/core';

@Directive({
    selector: '[ngxClipboard]',
})
export class ClipboardDirective implements OnDestroy {
    // eslint-disable-next-line @angular-eslint/no-input-rename
    @Input('ngxClipboard') public targetElm: HTMLInputElement;

    @Input() public cbContent: string;

    @Output() public cbOnSuccess: EventEmitter<any> = new EventEmitter<any>();

    @Output() public cbOnError: EventEmitter<any> = new EventEmitter<any>();
    constructor(
        private clipboardSrv: ClipboardService,
        private renderer: Renderer2
    ) {}

    public ngOnDestroy() {
        this.clipboardSrv.destroy();
    }

    @HostListener('click', ['$event.target'])
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    public onClick(button: ElementRef) {
        if (!this.clipboardSrv.isSupported) {
            this.handleResult(false, undefined);
        } else if (
            this.targetElm &&
            this.clipboardSrv.isTargetValid(this.targetElm)
        ) {
            this.handleResult(
                this.clipboardSrv.copyFromInputElement(
                    this.targetElm,
                    this.renderer
                ),
                this.targetElm.value
            );
        } else if (this.cbContent) {
            this.handleResult(
                this.clipboardSrv.copyFromContent(
                    this.cbContent,
                    this.renderer
                ),
                this.cbContent
            );
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
