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
import { AddWebhookComponent } from './add-webhook.component';
import { CUSTOM_ELEMENTS_SCHEMA, NO_ERRORS_SCHEMA } from '@angular/core';
import { SharedTestingModule } from '../../../../shared/shared.module';

describe('AddWebhookComponent', () => {
    let component: AddWebhookComponent;
    let fixture: ComponentFixture<AddWebhookComponent>;

    beforeEach(() => {
        TestBed.configureTestingModule({
            schemas: [CUSTOM_ELEMENTS_SCHEMA, NO_ERRORS_SCHEMA],
            imports: [SharedTestingModule],
            declarations: [AddWebhookComponent],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(AddWebhookComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
    it('should open modal and should be edit model', async () => {
        component.isEdit = true;
        component.isOpen = true;
        fixture.detectChanges();
        await fixture.whenStable();
        const body: HTMLElement =
            fixture.nativeElement.querySelector('.modal-body');
        expect(body).toBeTruthy();
        const title: HTMLElement =
            fixture.nativeElement.querySelector('.modal-title');
        expect(title.innerText).toEqual('WEBHOOK.EDIT_WEBHOOK');
    });
});
