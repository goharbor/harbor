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
import { ConfirmationDialogComponent } from './confirmation-dialog.component';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { ConfirmationTargets } from '../../entities/shared.const';
import { ConfirmationMessage } from '../../../base/global-confirmation-dialog/confirmation-message';
import { BatchInfo } from '../../../base/global-confirmation-dialog/confirmation-batch-message';
import { SharedTestingModule } from '../../shared.module';

describe('ConfirmationDialogComponent', () => {
    let comp: ConfirmationDialogComponent;
    let fixture: ComponentFixture<ConfirmationDialogComponent>;
    const deletionMessage: ConfirmationMessage = new ConfirmationMessage(
        'MEMBER.DELETION_TITLE',
        'MEMBER.DELETION_SUMMARY',
        'user1',
        {},
        ConfirmationTargets.CONFIG
    );

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [SharedTestingModule, BrowserAnimationsModule],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ConfirmationDialogComponent);
        comp = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(comp).toBeTruthy();
    });
    it('should open dialog and show right buttons', async () => {
        let message = deletionMessage;
        let buttons: HTMLElement;
        comp.open(message);
        fixture.detectChanges();
        await fixture.whenStable();
        buttons = fixture.nativeElement.querySelector('.modal-footer button');
        expect(buttons.textContent).toContain('BUTTON.CANCEL');
        message.buttons = 1;
        comp.open(message);
        fixture.detectChanges();
        await fixture.whenStable();
        buttons = fixture.nativeElement.querySelector('.modal-footer button');
        expect(buttons.textContent).toContain('BUTTON.NO');
        message.buttons = 3;
        comp.open(message);
        fixture.detectChanges();
        await fixture.whenStable();
        buttons = fixture.nativeElement.querySelector('.modal-footer button');
        expect(buttons.textContent).toContain('BUTTON.CLOSE');
    });
    it('check cancel and confirm function', async () => {
        let buttons: HTMLElement;
        comp.opened = true;
        comp.message = null;
        fixture.detectChanges();
        await fixture.whenStable();
        comp.cancel();
        fixture.detectChanges();
        await fixture.whenStable();
        buttons = fixture.nativeElement.querySelector('.modal-footer button');
        expect(buttons).toBeFalsy();
        comp.open(deletionMessage);
        fixture.detectChanges();
        await fixture.whenStable();
        comp.confirm();
        fixture.detectChanges();
        await fixture.whenStable();
        buttons = fixture.nativeElement.querySelector('.modal-footer button');
        expect(buttons).toBeFalsy();
    });
    it('check colorChange and toggleErrorTitle functions', () => {
        let batchInfo = new BatchInfo();
        const resultColor1: string = comp.colorChange(batchInfo);
        expect(resultColor1).toEqual('green');
        batchInfo.errorState = true;
        const resultColor2: string = comp.colorChange(batchInfo);
        expect(resultColor2).toEqual('red');
        batchInfo.loading = true;
        const resultColor3: string = comp.colorChange(batchInfo);
        expect(resultColor3).toEqual('#666');
        const errorSpan: HTMLSpanElement = document.createElement('span');
        errorSpan.style.display = 'none';
        comp.toggleErrorTitle(errorSpan);
        expect(errorSpan.style.display).toEqual('block');
        comp.toggleErrorTitle(errorSpan);
        expect(errorSpan.style.display).toEqual('none');
    });
});
