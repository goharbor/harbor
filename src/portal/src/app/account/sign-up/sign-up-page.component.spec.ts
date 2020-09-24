import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { UserService } from '../../user/user.service';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { MessageService } from '../../global-message/message.service';
import { RouterTestingModule } from '@angular/router/testing';
import { SignUpPageComponent } from './sign-up-page.component';
import { FormsModule } from '@angular/forms';
import { NewUserFormComponent } from '../../shared/new-user-form/new-user-form.component';
import { SessionService } from '../../shared/session.service';

describe('SignUpPageComponent', () => {
    let component: SignUpPageComponent;
    let fixture: ComponentFixture<SignUpPageComponent>;
    let fakeUserService = null;
    let fakeSessionService = null;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [SignUpPageComponent, NewUserFormComponent],
            imports: [
                FormsModule,
                RouterTestingModule,
                TranslateModule.forRoot()
            ],
            providers: [
                MessageService,
                TranslateService,
                { provide: UserService, useValue: fakeUserService },
                { provide: SessionService, useValue: fakeSessionService },
            ],
            schemas: [CUSTOM_ELEMENTS_SCHEMA]
        }).compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(SignUpPageComponent);
        component = fixture.componentInstance;
        component.newUserForm =
        TestBed.createComponent(NewUserFormComponent).componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
