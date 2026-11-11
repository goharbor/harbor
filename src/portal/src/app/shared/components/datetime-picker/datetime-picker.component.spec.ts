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
import { LOCALE_ID, NO_ERRORS_SCHEMA } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { DatePickerComponent } from './datetime-picker.component';
import { FormsModule } from '@angular/forms';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { DateValidatorDirective } from '../../directives/date-validator.directive';
import { registerLocaleData } from '@angular/common';
import locale_en from '@angular/common/locales/en';

describe('DatePickerComponent', () => {
    let component: DatePickerComponent;
    let fixture: ComponentFixture<DatePickerComponent>;
    registerLocaleData(locale_en, 'en-us');
    beforeEach(async () => {
        TestBed.overrideComponent(DatePickerComponent, {
            set: {
                providers: [
                    {
                        provide: LOCALE_ID,
                        useValue: 'en-us',
                    },
                ],
            },
        });
        await TestBed.configureTestingModule({
            imports: [FormsModule, TranslateModule.forRoot()],
            declarations: [DatePickerComponent, DateValidatorDirective],
            providers: [TranslateService],
            schemas: [NO_ERRORS_SCHEMA],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(DatePickerComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
