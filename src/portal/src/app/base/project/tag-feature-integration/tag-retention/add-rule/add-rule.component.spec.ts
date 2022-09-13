import { ComponentFixture, TestBed } from '@angular/core/testing';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { AddRuleComponent } from './add-rule.component';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import {
    BrowserAnimationsModule,
    NoopAnimationsModule,
} from '@angular/platform-browser/animations';
import { ClarityModule } from '@clr/angular';
import { FormsModule } from '@angular/forms';
import { RouterTestingModule } from '@angular/router/testing';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { TagRetentionService } from '../tag-retention.service';
import { ErrorHandler } from '../../../../../shared/units/error-handler';
import { InlineAlertComponent } from '../../../../../shared/components/inline-alert/inline-alert.component';
describe('AddRuleComponent', () => {
    let component: AddRuleComponent;
    let fixture: ComponentFixture<AddRuleComponent>;
    const mockTagRetentionService = {};

    beforeEach(() => {
        TestBed.configureTestingModule({
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
            imports: [
                BrowserAnimationsModule,
                ClarityModule,
                TranslateModule.forRoot(),
                FormsModule,
                RouterTestingModule,
                NoopAnimationsModule,
                HttpClientTestingModule,
            ],
            declarations: [AddRuleComponent, InlineAlertComponent],
            providers: [
                TranslateService,
                ErrorHandler,
                {
                    provide: TagRetentionService,
                    useValue: mockTagRetentionService,
                },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(AddRuleComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
