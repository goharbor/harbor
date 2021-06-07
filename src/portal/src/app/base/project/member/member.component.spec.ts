import { ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';
import { MemberComponent } from './member.component';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { ActivatedRoute } from "@angular/router";
import { MessageHandlerService } from "../../../shared/services/message-handler.service";
import { SessionService } from "../../../shared/services/session.service";
import { MemberService } from "./member.service";
import { AppConfigService } from "../../../services/app-config.service";
import { of } from 'rxjs';
import { OperationService } from "../../../shared/components/operation/operation.service";
import { UserPermissionService } from "../../../shared/services";
import { ErrorHandler } from "../../../shared/units/error-handler";
import { ConfirmationDialogService } from "../../global-confirmation-dialog/confirmation-dialog.service";
import { SharedTestingModule } from "../../../shared/shared.module";
import { HttpHeaders, HttpResponse } from "@angular/common/http";
import { Registry } from "../../../../../ng-swagger-gen/models/registry";

describe('MemberComponent', () => {
    let component: MemberComponent;
    let fixture: ComponentFixture<MemberComponent>;
    const mockMemberService = {
        getUsersNameList: () => {
            return of([]);
        },
        listProjectMembersResponse: () => {
            const response: HttpResponse<Array<Registry>> = new HttpResponse<Array<Registry>>({
                headers: new HttpHeaders({'x-total-count': '0'}),
                body: []
            });
            return of(response);
        },
        updateProjectMember: () => {
            return of(null);
        },
        deleteProjectMember: () => {
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


    beforeEach(waitForAsync(() => {
        TestBed.configureTestingModule({
            schemas: [
                CUSTOM_ELEMENTS_SCHEMA
            ],
            imports: [
                SharedTestingModule
            ],
            declarations: [MemberComponent],
            providers: [
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
                                parent: {
                                    params: { id: 1 }
                                }
                            },
                        }
                    }
                }

            ]
        }).compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(MemberComponent);
        component = fixture.componentInstance;
        component.loading = true;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
