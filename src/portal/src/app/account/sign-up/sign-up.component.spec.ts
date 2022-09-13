import { ComponentFixture, TestBed } from '@angular/core/testing';
import { SignUpComponent } from './sign-up.component';
import { SessionService } from '../../shared/services/session.service';
import { UserService } from '../../base/left-side-nav/user/user.service';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { NewUserFormComponent } from '../../shared/components/new-user-form/new-user-form.component';
import { of } from 'rxjs';
import { InlineAlertComponent } from '../../shared/components/inline-alert/inline-alert.component';
import { SharedTestingModule } from '../../shared/shared.module';

describe('SignUpComponent', () => {
    let component: SignUpComponent;
    let fixture: ComponentFixture<SignUpComponent>;
    let fakeSessionService = {
        checkUserExisting: () => of(true),
    };
    let fakeUserService = {
        addUser: () => of(null),
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
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [
                SignUpComponent,
                NewUserFormComponent,
                InlineAlertComponent,
            ],
            imports: [SharedTestingModule],
            providers: [
                { provide: SessionService, useValue: fakeSessionService },
                { provide: UserService, useValue: fakeUserService },
            ],
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
        }).compileComponents();
    });

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
        const closeBtn: HTMLButtonElement =
            fixture.nativeElement.querySelector('#close-btn');
        expect(closeBtn).toBeTruthy();
        closeBtn.dispatchEvent(new Event('click'));
        await fixture.whenStable();
        const closeBtn1: HTMLButtonElement =
            fixture.nativeElement.querySelector('#close-btn');
        expect(closeBtn1).toBeNull();
    });
    it('should create new user', async () => {
        component.open();
        component.getNewUser = () => mockUser;
        await fixture.whenStable();
        const createBtn = fixture.nativeElement.querySelector('#sign-up');
        createBtn.dispatchEvent(new Event('click'));

        await fixture.whenStable();
        const closeBtn1: HTMLButtonElement =
            fixture.nativeElement.querySelector('#close-btn');
        expect(closeBtn1).toBeNull();
    });
});
