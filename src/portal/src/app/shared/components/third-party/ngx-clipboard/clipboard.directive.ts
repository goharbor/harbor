// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
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
