import { ComponentFixture, TestBed } from '@angular/core/testing';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { NewUserFormComponent } from './new-user-form.component';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import {
    BrowserAnimationsModule,
    NoopAnimationsModule,
} from '@angular/platform-browser/animations';
import { ClarityModule } from '@clr/angular';
import { FormsModule } from '@angular/forms';
import { RouterTestingModule } from '@angular/router/testing';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { SessionService } from '../../services/session.service';

describe('NewUserFormComponent', () => {
    let component: NewUserFormComponent;
    let fixture: ComponentFixture<NewUserFormComponent>;
    const mockSessionService = {
        getCurrentUser: () => {},
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
            declarations: [NewUserFormComponent],
            providers: [
                { provide: SessionService, useValue: mockSessionService },
                TranslateService,
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(NewUserFormComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
