import { ComponentFixture, TestBed } from '@angular/core/testing';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import {
    BrowserAnimationsModule,
    NoopAnimationsModule,
} from '@angular/platform-browser/animations';
import { ClarityModule } from '@clr/angular';
import { FormsModule } from '@angular/forms';
import { RouterTestingModule } from '@angular/router/testing';
import { of } from 'rxjs';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { ConfirmationDialogService } from './confirmation-dialog.service';
import { GlobalConfirmationDialogComponent } from './global-confirmation-dialog.component';

describe('ConfirmationDialogComponent', () => {
    let component: GlobalConfirmationDialogComponent;
    let fixture: ComponentFixture<GlobalConfirmationDialogComponent>;
    const mockConfirmationDialogService = {
        confirmationAnnouced$: of({
            title: 'title',
            message: 'title',
            param: 'AAA',
        }),
        cancel: () => {},
        confirm: () => {},
    };
    beforeEach(async () => {
        await TestBed.configureTestingModule({
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
            declarations: [GlobalConfirmationDialogComponent],
            providers: [
                TranslateService,
                {
                    provide: ConfirmationDialogService,
                    useValue: mockConfirmationDialogService,
                },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(GlobalConfirmationDialogComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
