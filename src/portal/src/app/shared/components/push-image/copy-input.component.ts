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
import { Component, Input, Output, EventEmitter } from '@angular/core';
import { ClipboardService } from '../third-party/ngx-clipboard';

export const enum CopyStatus {
    NORMAL,
    SUCCESS,
    ERROR,
}

@Component({
    selector: 'hbr-copy-input',
    templateUrl: './copy-input.coponent.html',
    styleUrls: ['./push-image.scss'],
})
export class CopyInputComponent {
    @Input() inputSize: number = 40;
    @Input() headerTitle: string = 'Copy Input';
    @Input() defaultValue: string = 'N/A';
    @Input() iconMode: boolean = false;
    @Input() linkMode: boolean = false;
    @Input() linkName: string = 'N/A';

    state: CopyStatus = CopyStatus.NORMAL;

    // eslint-disable-next-line @angular-eslint/no-output-on-prefix
    @Output() onCopySuccess: EventEmitter<any> = new EventEmitter<any>();
    // eslint-disable-next-line @angular-eslint/no-output-on-prefix
    @Output() onCopyError: EventEmitter<any> = new EventEmitter<any>();

    constructor(private clipboardSrv: ClipboardService) {}
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
        return (
            this.state === CopyStatus.SUCCESS &&
            this.clipboardSrv?.getCurrentCopiedText() === this.defaultValue
        );
    }

    public get hasCopyError(): boolean {
        return this.state === CopyStatus.ERROR;
    }
}
