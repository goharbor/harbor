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
