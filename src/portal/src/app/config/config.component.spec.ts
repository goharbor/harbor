import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { SessionService } from '../shared/session.service';
import { ConfirmationDialogService } from '../shared/confirmation-dialog/confirmation-dialog.service';
import { MessageHandlerService } from '../shared/message-handler/message-handler.service';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { ClarityModule } from "@clr/angular";
import { AppConfigService } from '../app-config.service';
import { ConfigurationService } from './config.service';
import { ConfigurationComponent } from './config.component';
import { of } from 'rxjs';
import { ConfirmationAcknowledgement } from '../shared/confirmation-dialog/confirmation-state-message';
import { ConfirmationState, ConfirmationTargets } from '../shared/shared.const';
import { Configuration } from '../../lib/components/config/config';

describe('ConfigurationComponent', () => {
    let component: ConfigurationComponent;
    let fixture: ComponentFixture<ConfigurationComponent>;
    let confirmationConfirmFlag = true;
    let confirmationConfirm = () => {
        return confirmationConfirmFlag ? of(new ConfirmationAcknowledgement(ConfirmationState.CONFIRMED, {}, ConfirmationTargets.CONFIG))
        : of(new ConfirmationAcknowledgement(ConfirmationState.CONFIRMED
            , {change: { email_password: 'Harbor12345' }, tabId: '1'}, ConfirmationTargets.CONFIG_TAB));
    };
    let fakeConfirmationDialogService = {
        confirmationConfirm$: confirmationConfirm()
    };
     let mockConfigurationService = {
        getConfiguration: () => of(new Configuration()),
        confirmationConfirm$: of(new ConfirmationAcknowledgement(ConfirmationState.CONFIRMED, {}, ConfirmationTargets.CONFIG))
     };
     let fakeSessionService = {
        getCurrentUser: function () {
            return {
                sysadmin_flag: true,
                admin_role_in_auth: true,
                user_id: 1,
                username: 'admin',
                email: "",
                realname: "admin",
                role_name: "admin",
                role_id: 1,
                comment: "string",
            };
        },
        updateAccountSettings: () => of(null),
        renameAdmin: () => of(null),
    };
    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [
                TranslateModule.forRoot(),
                ClarityModule
            ],
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
            declarations: [ConfigurationComponent],
            providers: [
                TranslateService,
                {
                    provide: SessionService, useValue: {
                        getCurrentUser: function () {
                            return "admin";
                        }
                    }
                },
                { provide: ConfirmationDialogService, useValue: fakeConfirmationDialogService },
                { provide: SessionService, useValue: fakeSessionService },
                { provide: MessageHandlerService, useValue: null },
                {
                    provide: AppConfigService, useValue: {
                        getConfig: function () {
                            return { has_ca_root: true };
                        }
                    }
                },
                {
                    provide: ConfigurationService, useValue: mockConfigurationService
                }
            ]
        }).compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(ConfigurationComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
    it('should reset part of allConfig ', async () => {
        confirmationConfirmFlag = false;
        component.originalCopy.email_password.value = 'Harbor12345';
        component.reset({
            email_password: {
                value: 'Harbor12345',
                editable: true
            }
        });
        await fixture.whenStable();
        expect(component.allConfig.email_password.value).toEqual('Harbor12345');
    });
});
