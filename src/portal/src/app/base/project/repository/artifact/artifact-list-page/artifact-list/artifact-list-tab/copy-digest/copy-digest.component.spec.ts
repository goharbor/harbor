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
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { CopyDigestComponent } from './copy-digest.component';
import { SharedTestingModule } from '../../../../../../../../shared/shared.module';
import { Clipboard } from '@angular/cdk/clipboard';

describe('CopyDigestComponent', () => {
    let component: CopyDigestComponent;
    let fixture: ComponentFixture<CopyDigestComponent>;
    let clipboardSpy: jasmine.SpyObj<Clipboard>;

    beforeEach(async () => {
        // Create a spy for the Angular CDK Clipboard service
        clipboardSpy = jasmine.createSpyObj('Clipboard', ['copy']);

        await TestBed.configureTestingModule({
            declarations: [CopyDigestComponent],
            imports: [SharedTestingModule],
            // Provide the Clipboard spy instead of the real service
            providers: [
                { provide: Clipboard, useValue: clipboardSpy },
            ],
        }).compileComponents();

        fixture = TestBed.createComponent(CopyDigestComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should show right digest', async () => {
        const digest: string = 'sha256@test';
        component.showDigestId(digest);
        fixture.detectChanges();
        await fixture.whenStable();
        const textArea: HTMLTextAreaElement =
            fixture.nativeElement.querySelector(`textarea`);
        expect(textArea.textContent).toEqual(digest);
    });

    // Test that copyToClipboard calls the CDK Clipboard service with correct text
    it('should copy text to clipboard successfully', () => {
        clipboardSpy.copy.and.returnValue(true);
        component.copyToClipboard('sha256@test');
        expect(clipboardSpy.copy).toHaveBeenCalledWith('sha256@test');
        // On success modal should close and copyFailed should be false
        expect(component.copyFailed).toBeFalse();
        expect(component.showTagManifestOpened).toBeFalse();
    });

    // Test that copyFailed is set to true when clipboard copy fails
    it('should handle clipboard copy failure', () => {
        clipboardSpy.copy.and.returnValue(false);
        component.copyToClipboard('sha256@test');
        expect(clipboardSpy.copy).toHaveBeenCalledWith('sha256@test');
        // On failure copyFailed should be true
        expect(component.copyFailed).toBeTrue();
    });

    // Test that showDigestId opens the modal and resets copyFailed
    it('should open modal and reset state when showDigestId is called', () => {
        component.copyFailed = true;
        component.showDigestId('sha256@test');
        expect(component.showTagManifestOpened).toBeTrue();
        expect(component.digestId).toEqual('sha256@test');
        // copyFailed should reset when modal opens
        expect(component.copyFailed).toBeFalse();
    });
});