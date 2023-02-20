import { ComponentFixture, TestBed } from '@angular/core/testing';
import { MemberComponent } from './member.component';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { MessageHandlerService } from '../../../shared/services/message-handler.service';
import { SessionService } from '../../../shared/services/session.service';
import { MemberService } from '../../../../../ng-swagger-gen/services/member.service';
import { AppConfigService } from '../../../services/app-config.service';
import { of } from 'rxjs';
import { OperationService } from '../../../shared/components/operation/operation.service';
import { UserPermissionService } from '../../../shared/services';
import { ErrorHandler } from '../../../shared/units/error-handler';
import { ConfirmationDialogService } from '../../global-confirmation-dialog/confirmation-dialog.service';
import { SharedTestingModule } from '../../../shared/shared.module';
import { HttpHeaders, HttpResponse } from '@angular/common/http';
import { Registry } from '../../../../../ng-swagger-gen/models/registry';
import { ProjectMemberEntity } from '../../../../../ng-swagger-gen/models/project-member-entity';
import { delay } from 'rxjs/operators';

describe('MemberComponent', () => {
    let component: MemberComponent;
    let fixture: ComponentFixture<MemberComponent>;
    const mockedMembers: ProjectMemberEntity[] = [
        {
            id: 1,
            project_id: 1,
            entity_name: 'test1',
            role_name: 'projectAdmin',
            role_id: 1,
            entity_id: 1,
            entity_type: 'u',
        },
        {
            id: 2,
            project_id: 1,
            entity_name: 'test2',
            role_name: 'projectAdmin',
            role_id: 1,
            entity_id: 2,
            entity_type: 'u',
        },
    ];
    const mockMemberService = {
        getUsersNameList: () => {
            return of([]);
        },
        listProjectMembersResponse: () => {
            const response: HttpResponse<Array<Registry>> = new HttpResponse<
                Array<Registry>
            >({
                headers: new HttpHeaders({ 'x-total-count': '2' }),
                body: mockedMembers,
            });
            return of(response).pipe(delay(0));
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
                user_id: 1,
            });
        },
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
        publishInfo: () => {},
    };
    const mockMessageHandlerService = {
        handleError: () => {},
    };
    const mockConfirmationDialogService = {
        openComfirmDialog: () => {},
        confirmationConfirm$: of({
            state: 1,
            source: 2,
        }),
    };
    const mockUserPermissionService = {
        getPermission() {
            return of(true);
        },
    };
    const mockErrorHandler = {
        error() {},
    };

    beforeEach(() => {
        TestBed.configureTestingModule({
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
            imports: [SharedTestingModule],
            declarations: [MemberComponent],
            providers: [
                { provide: MemberService, useValue: mockMemberService },
                {
                    provide: MessageHandlerService,
                    useValue: mockMessageHandlerService,
                },
                {
                    provide: ConfirmationDialogService,
                    useValue: mockConfirmationDialogService,
                },
                { provide: SessionService, useValue: mockSessionService },
                { provide: OperationService, useValue: mockOperationService },
                { provide: AppConfigService, useValue: mockAppConfigService },
                {
                    provide: UserPermissionService,
                    useValue: mockUserPermissionService,
                },
                { provide: ErrorHandler, useValue: mockErrorHandler },
                {
                    provide: ActivatedRoute,
                    useValue: {
                        RouterparamMap: of({ get: key => 'value' }),
                        snapshot: {
                            parent: {
                                parent: {
                                    params: { id: 1 },
                                },
                            },
                        },
                    },
                },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(MemberComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
    it('should render member list', async () => {
        fixture.autoDetectChanges(true);
        await fixture.whenStable();
        const rows = fixture.nativeElement.querySelectorAll('clr-dg-row');
        expect(rows.length).toEqual(2);
    });
});
