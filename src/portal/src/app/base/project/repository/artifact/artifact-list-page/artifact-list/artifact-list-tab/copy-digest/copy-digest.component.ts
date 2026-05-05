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
import { Component, ElementRef, ViewChild } from '@angular/core';

@Component({
    selector: 'app-copy-digest',
    templateUrl: './copy-digest.component.html',
    styleUrls: ['./copy-digest.component.scss'],
})
export class CopyDigestComponent {
    showTagManifestOpened: boolean = false;
    digestId: string;
    @ViewChild('digestTarget') textInput: ElementRef;
    copyFailed: boolean = false;
    constructor() {}
    onSuccess($event: any): void {
        this.copyFailed = false;
        // Directly close dialog
        this.showTagManifestOpened = false;
    }

    onError($event: any): void {
        // Show error
        this.copyFailed = true;
        // Select all text
        if (this.textInput) {
            this.textInput.nativeElement.select();
        }
    }
    showDigestId(digest: string) {
        this.digestId = digest;
        this.showTagManifestOpened = true;
        this.copyFailed = false;
    }
}
