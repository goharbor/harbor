import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { MemberComponent } from './member.component';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { BrowserAnimationsModule, NoopAnimationsModule } from '@angular/platform-browser/animations';
import { ClarityModule } from '@clr/angular';
import { FormsModule } from '@angular/forms';
import { RouterTestingModule } from '@angular/router/testing';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { ActivatedRoute } from "@angular/router";
import { MessageHandlerService } from "../../shared/message-handler/message-handler.service";
import { ConfirmationDialogService } from "../../shared/confirmation-dialog/confirmation-dialog.service";
import { SessionService } from "../../shared/session.service";
import { MemberService } from "./member.service";
import { AppConfigService } from "../../app-config.service";
import { of } from 'rxjs';
import { OperationService } from "../../../lib/components/operation/operation.service";
import { UserPermissionService } from "../../../lib/services";
import { ErrorHandler } from "../../../lib/utils/error-handler";

describe('MemberComponent', () => {
    let component: MemberComponent;
    let fixture: ComponentFixture<MemberComponent>;
    const mockMemberService = {
        getUsersNameList: () => {
            return of([]);
        },
        listMembers: () => {
            return of([]);
        },
        changeMemberRole: () => {
            return of(null);
        },
        deleteMember: () => {
            return of(null);
        },
    };
    const mockSessionService = {
        getCurrentUser: () => {
            return of({
                user_id: 1
            });
        }
    };
    const mockAppConfigService = {
        isLdapMode: () => {
            return false;
        },
        isHttpAuthMode: () => {
            return false;
        },
        isOidcMode: () => {
            return true;
        },

    };
    const mockOperationService = {
        publishInfo: () => { }
    };
    const mockMessageHandlerService = {
        handleError: () => { }
    };
    const mockConfirmationDialogService = {
        openComfirmDialog: () => { },
        confirmationConfirm$:  of(
                {
                    state: 1,
                    source: 2,
                }
            )
    };
    const mockUserPermissionService = {
        getPermission() {
          return of(true);
        }
      };
    const mockErrorHandler = {
        error() { }
      };


    beforeEach(async(() => {
        TestBed.configureTestingModule({
            schemas: [
                CUSTOM_ELEMENTS_SCHEMA
            ],
            imports: [
                BrowserAnimationsModule,
                ClarityModule,
                TranslateModule.forRoot(),
                FormsModule,
                RouterTestingModule,
                NoopAnimationsModule,
                HttpClientTestingModule
            ],
            declarations: [MemberComponent],
            providers: [
                TranslateService,
                { provide: MemberService, useValue: mockMemberService },
                { provide: MessageHandlerService, useValue: mockMessageHandlerService },
                { provide: ConfirmationDialogService, useValue: mockConfirmationDialogService },
                { provide: SessionService, useValue: mockSessionService },
                { provide: OperationService, useValue: mockOperationService },
                { provide: AppConfigService, useValue: mockAppConfigService },
                { provide: UserPermissionService, useValue: mockUserPermissionService },
                { provide: ErrorHandler, useValue: mockErrorHandler },
                {
                    provide: ActivatedRoute, useValue: {
                        RouterparamMap: of({ get: (key) => 'value' }),
                        snapshot: {
                            parent: {
                                params: { id: 1 }
                            },
                            data: 1
                        }
                    }
                }

            ]
        }).compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(MemberComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
