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
