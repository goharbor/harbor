import { ComponentFixture, TestBed } from '@angular/core/testing';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { UserService } from './user.service';
import { MessageHandlerService } from '../../../shared/services/message-handler.service';
import { AppConfigService } from '../../../services/app-config.service';
import { SessionService } from '../../../shared/services/session.service';
import { UserComponent } from './user.component';
import { OperationService } from '../../../shared/components/operation/operation.service';
import { ConfirmationDialogService } from '../../global-confirmation-dialog/confirmation-dialog.service';
import { SharedTestingModule } from '../../../shared/shared.module';

describe('UserComponent', () => {
    let component: UserComponent;
    let fixture: ComponentFixture<UserComponent>;
    let fakeSessionService = null;
    let fakeAppConfigService = {
        getConfig: function () {
            return {
                auth_mode: 'ldap_auth',
            };
        },
    };
    let fakeUserService = null;
    let fakeMessageHandlerService = {
        handleError: function () {},
    };

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [UserComponent],
            imports: [SharedTestingModule],
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
            providers: [
                ConfirmationDialogService,
                OperationService,
                {
                    provide: MessageHandlerService,
                    useValue: fakeMessageHandlerService,
                },
                { provide: UserService, useValue: fakeUserService },
                { provide: SessionService, useValue: fakeSessionService },
                { provide: AppConfigService, useValue: fakeAppConfigService },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(UserComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
