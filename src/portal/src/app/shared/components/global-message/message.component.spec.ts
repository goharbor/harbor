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
import { ElementRef } from '@angular/core';
import { MessageComponent } from './message.component';
import { SharedTestingModule } from '../../shared.module';

describe('MessageComponent', () => {
    let component: MessageComponent;
    let fixture: ComponentFixture<MessageComponent>;
    let fakeElementRef = null;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [MessageComponent],
            providers: [{ provide: ElementRef, useValue: fakeElementRef }],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(MessageComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('onAlertClosedChange should close message when closed is true', () => {
        component.globalMessageOpened = true;
        component.onAlertClosedChange(true);
        expect(component.globalMessageOpened).toBeFalse();
    });

    it('onAlertClosedChange should not close message when closed is false', () => {
        component.globalMessageOpened = true;
        component.onAlertClosedChange(false);
        expect(component.globalMessageOpened).toBeTrue();
    });
});
