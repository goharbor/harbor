import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { ClarityModule } from '@clr/angular';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { UserService } from './user.service';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { ConfirmationDialogService } from '../shared/confirmation-dialog/confirmation-dialog.service';
import { MessageHandlerService } from '../shared/message-handler/message-handler.service';
import { AppConfigService } from '../app-config.service';
import { SessionService } from '../shared/session.service';
import { UserComponent } from './user.component';
import { OperationService } from "../../lib/components/operation/operation.service";

describe('UserComponent', () => {
    let component: UserComponent;
    let fixture: ComponentFixture<UserComponent>;
    let fakeSessionService = null;
    let fakeAppConfigService = {
        getConfig: function () {
            return {
                auth_mode: 'ldap_auth'
            };
        }
    };
    let fakeUserService = null;
    let fakeMessageHandlerService = {
        handleError: function () { }
    };

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [UserComponent],
            imports: [
                ClarityModule,
                TranslateModule.forRoot(),
                HttpClientTestingModule
            ],
            schemas: [
                CUSTOM_ELEMENTS_SCHEMA
            ],
            providers: [
                TranslateService,
                ConfirmationDialogService,
                OperationService,
                { provide: MessageHandlerService, useValue: fakeMessageHandlerService },
                { provide: UserService, useValue: fakeUserService },
                { provide: SessionService, useValue: fakeSessionService },
                { provide: AppConfigService, useValue: fakeAppConfigService }
            ]
        }).compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(UserComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
