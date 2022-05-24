import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ActivatedRoute } from '@angular/router';
import { MessageHandlerService } from '../../../../shared/services/message-handler.service';
import { of } from 'rxjs';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { delay } from 'rxjs/operators';
import { ConfirmationDialogComponent } from '../../../../shared/components/confirmation-dialog';
import {
    ProjectService,
    UserPermissionService,
} from '../../../../shared/services';
import { PolicyComponent } from './policy.component';
import { PreheatService } from '../../../../../../ng-swagger-gen/services/preheat.service';
import { AddP2pPolicyComponent } from '../add-p2p-policy/add-p2p-policy.component';
import { PreheatPolicy } from '../../../../../../ng-swagger-gen/models/preheat-policy';
import { Execution } from '../../../../../../ng-swagger-gen/models/execution';
import { ProviderUnderProject } from '../../../../../../ng-swagger-gen/models/provider-under-project';
import { P2pProviderService } from '../p2p-provider.service';
import { SessionService } from '../../../../shared/services/session.service';
import { AppConfigService } from '../../../../services/app-config.service';
import { ErrorHandler } from '../../../../shared/units/error-handler';
import { InlineAlertComponent } from '../../../../shared/components/inline-alert/inline-alert.component';
import { SharedTestingModule } from '../../../../shared/shared.module';
import { HttpHeaders, HttpResponse } from '@angular/common/http';
import { Registry } from '../../../../../../ng-swagger-gen/models/registry';

describe('PolicyComponent', () => {
    let component: PolicyComponent;
    let fixture: ComponentFixture<PolicyComponent>;
    const mockMessageHandlerService = {
        handleError: () => {},
    };
    const providers: ProviderUnderProject[] = [
        {
            id: 1,
            provider: 'Kraken',
        },
    ];
    const mockActivatedRoute = {
        snapshot: {
            parent: {
                parent: {
                    params: { id: 1 },
                    data: {
                        projectResolver: {
                            name: 'library',
                            metadata: {
                                prevent_vul: 'true',
                                enable_content_trust: 'true',
                                severity: 'none',
                            },
                        },
                    },
                },
            },
        },
    };
    const mockUserPermissionService = {
        getPermission() {
            return of(true).pipe(delay(0));
        },
    };
    const mockedSessionService = {
        getCurrentUser() {
            return {
                has_admin_role: true,
            };
        },
    };

    const policy1: PreheatPolicy = {
        id: 1,
        name: 'policy1',
        provider_id: 1,
        enabled: true,
        filters:
            '[{"type":"repository","value":"**"},{"type":"tag","value":"**"},{"type":"vulnerability","value":2}]',
        trigger: '{"type":"manual","trigger_setting":{"cron":""}}',
        creation_time: '2020-01-02T15:04:05',
        description: 'policy1',
    };

    const policy2: PreheatPolicy = {
        id: 2,
        name: 'policy2',
        provider_id: 2,
        enabled: false,
        filters:
            '[{"type":"repository","value":"**"},{"type":"tag","value":"**"},{"type":"vulnerability","value":2}]',
        trigger: '{"type":"manual","trigger_setting":{"cron":""}}',
        creation_time: '2020-01-02T15:04:05',
        description: 'policy2',
    };
    const execution: Execution = {
        id: 1,
        vendor_id: 1,
        status: 'Success',
        trigger: 'Manual',
        start_time: new Date().toUTCString(),
    };
    const mockedAppConfigService = {
        getConfig() {
            return {
                with_notary: true,
            };
        },
    };

    const mockPreheatService = {
        ListPoliciesResponse: () => {
            const response: HttpResponse<Array<Registry>> = new HttpResponse<
                Array<Registry>
            >({
                headers: new HttpHeaders({
                    'x-total-count': [policy1, policy2].length.toString(),
                }),
                body: [policy1, policy2],
            });
            return of(response).pipe(delay(0));
        },
        ListProvidersUnderProject() {
            return of(providers).pipe(delay(0));
        },
        ListExecutionsResponse() {
            return of([execution]).pipe(delay(0));
        },
    };
    const mockedProjectService = {
        getProject() {
            return of({
                name: 'library',
                metadata: {
                    prevent_vul: 'true',
                    enable_content_trust: 'true',
                    severity: 'none',
                },
            });
        },
    };
    beforeEach(() => {
        TestBed.configureTestingModule({
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
            imports: [SharedTestingModule],
            declarations: [
                PolicyComponent,
                AddP2pPolicyComponent,
                InlineAlertComponent,
                ConfirmationDialogComponent,
            ],
            providers: [
                P2pProviderService,
                ErrorHandler,
                { provide: PreheatService, useValue: mockPreheatService },
                {
                    provide: MessageHandlerService,
                    useValue: mockMessageHandlerService,
                },
                { provide: ActivatedRoute, useValue: mockActivatedRoute },
                {
                    provide: UserPermissionService,
                    useValue: mockUserPermissionService,
                },
                { provide: SessionService, useValue: mockedSessionService },
                { provide: AppConfigService, useValue: mockedAppConfigService },
                { provide: ProjectService, useValue: mockedProjectService },
            ],
        }).compileComponents();
    });

    beforeEach(async () => {
        fixture = TestBed.createComponent(PolicyComponent);
        component = fixture.componentInstance;
        fixture.autoDetectChanges(true);
    });

    it('should create', async () => {
        expect(component).toBeTruthy();
    });
    it('should get policy list', async () => {
        await fixture.whenStable();
        const rows = fixture.nativeElement.getElementsByTagName('clr-dg-row');
        expect(rows.length).toEqual(2);
    });
    it('should open modal and is add model', async () => {
        await fixture.whenStable();
        const addButton: HTMLButtonElement =
            fixture.nativeElement.querySelector('#new-policy');
        addButton.click();
        await fixture.whenStable();
        const modalBody: HTMLDivElement =
            fixture.nativeElement.querySelector('.modal-body');
        expect(modalBody).toBeTruthy();
        const title: HTMLElement =
            fixture.nativeElement.querySelector('.modal-title');
        expect(title.innerText).toEqual('P2P_PROVIDER.ADD_POLICY');
    });
    it('should open modal and is edit model', async () => {
        component.selectedRow = policy1;
        await fixture.whenStable();
        const action: HTMLSpanElement =
            fixture.nativeElement.querySelector('#action-policy');
        action.click();
        await fixture.whenStable();
        const edit: HTMLSpanElement =
            fixture.nativeElement.querySelector('#edit-policy');
        edit.click();
        await fixture.whenStable();
        const modalBody: HTMLDivElement =
            fixture.nativeElement.querySelector('.modal-body');
        expect(modalBody).toBeTruthy();
        const nameInput: HTMLInputElement =
            fixture.nativeElement.querySelector('#name');
        expect(nameInput.value).toEqual('policy1');
    });
});
