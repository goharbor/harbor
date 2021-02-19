import { waitForAsync, ComponentFixture, TestBed } from '@angular/core/testing';
import { ClarityModule } from "@clr/angular";
import { SignUpComponent } from './sign-up.component';
import { SessionService } from '../../shared/services/session.service';
import { UserService } from '../../base/left-side-nav/user/user.service';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { NewUserFormComponent } from '../../shared/components/new-user-form/new-user-form.component';
import { FormsModule } from '@angular/forms';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { of } from 'rxjs';
import { ErrorHandler } from '../../shared/units/error-handler';
import { InlineAlertComponent } from "../../shared/components/inline-alert/inline-alert.component";

describe('SignUpComponent', () => {
    let component: SignUpComponent;
    let fixture: ComponentFixture<SignUpComponent>;
    let fakeSessionService = {
        checkUserExisting: () => of(true)
    };
    let fakeUserService = {
        addUser: () => of(null)
    };
    const mockUser = {
        user_id: 1,
        username: 'string',
        realname: 'string',
        email: 'string',
        password: 'string',
        comment: 'string',
        deleted: false,
        role_name: 'string',
        role_id: 1,
        has_admin_role: true,
        reset_uuid: 'string',
        creation_time: 'string',
        update_time: 'string',
    };
    beforeEach(waitForAsync(() => {
        TestBed.configureTestingModule({
            declarations: [SignUpComponent, NewUserFormComponent, InlineAlertComponent],
            imports: [
                FormsModule,
                ClarityModule,
                TranslateModule.forRoot(),
                BrowserAnimationsModule
            ],
            providers: [
                TranslateService,
                { provide: SessionService, useValue: fakeSessionService },
                { provide: UserService, useValue: fakeUserService },
                ErrorHandler
            ],
            schemas: [CUSTOM_ELEMENTS_SCHEMA]
        }).compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(SignUpComponent);
        component = fixture.componentInstance;
        component.newUserForm =
            TestBed.createComponent(NewUserFormComponent).componentInstance;
        component.inlineAlert =
            TestBed.createComponent(InlineAlertComponent).componentInstance;
        component.opened = true;
        fixture.autoDetectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should close when no form change', async () => {
        component.open();
        await fixture.whenStable();
        const closeBtn: HTMLButtonElement = fixture.nativeElement.querySelector("#close-btn");
        expect(closeBtn).toBeTruthy();
        closeBtn.dispatchEvent(new Event('click'));
        await fixture.whenStable();
        const closeBtn1: HTMLButtonElement = fixture.nativeElement.querySelector("#close-btn");
        expect(closeBtn1).toBeNull();

    });
    it('should create new user', async () => {
        component.open();
        component.getNewUser = () => mockUser;
        await fixture.whenStable();
        const createBtn = fixture.nativeElement.querySelector('#sign-up');
        createBtn.dispatchEvent(new Event('click'));

        await fixture.whenStable();
        const closeBtn1: HTMLButtonElement = fixture.nativeElement.querySelector("#close-btn");
        expect(closeBtn1).toBeNull();

    });
});
