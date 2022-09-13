import { ComponentFixture, TestBed } from '@angular/core/testing';
import { WebhookComponent } from './webhook.component';
import { ActivatedRoute } from '@angular/router';
import { ProjectWebhookService } from './webhook.service';
import { MessageHandlerService } from '../../../shared/services/message-handler.service';
import { of } from 'rxjs';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { delay } from 'rxjs/operators';
import { AddWebhookFormComponent } from './add-webhook-form/add-webhook-form.component';
import { AddWebhookComponent } from './add-webhook/add-webhook.component';
import { ConfirmationDialogComponent } from '../../../shared/components/confirmation-dialog';
import { UserPermissionService } from '../../../shared/services';
import { InlineAlertComponent } from '../../../shared/components/inline-alert/inline-alert.component';
import { SharedTestingModule } from '../../../shared/shared.module';
import { WebhookPolicy } from '../../../../../ng-swagger-gen/models/webhook-policy';
import { WebhookService } from '../../../../../ng-swagger-gen/services/webhook.service';
import { HttpHeaders, HttpResponse } from '@angular/common/http';
import { Registry } from '../../../../../ng-swagger-gen/models/registry';

describe('WebhookComponent', () => {
    let component: WebhookComponent;
    let fixture: ComponentFixture<WebhookComponent>;
    const mockMessageHandlerService = {
        handleError: () => {},
    };
    const mockedMetadata = {
        event_type: [
            'projectQuota',
            'pullImage',
            'scanningFailed',
            'uploadChart',
            'deleteChart',
            'downloadChart',
            'scanningCompleted',
            'pushImage',
            'deleteImage',
        ],
        notify_type: ['http', 'slack'],
    };
    const mockedWehook: WebhookPolicy = {
        id: 1,
        project_id: 1,
        name: 'test',
        description: 'just a test webhook',
        targets: [
            {
                address: 'https://test.com',
                type: 'http',
                auth_header: null,
                skip_cert_verify: true,
            },
        ],
        event_types: ['projectQuota'],
        creator: null,
        creation_time: null,
        update_time: null,
        enabled: true,
    };
    const mockProjectWebhookService = {
        eventTypeToText(eventType: string) {
            return eventType;
        },
    };
    const mockedWebhookService = {
        GetSupportedEventTypes() {
            return of(mockedMetadata).pipe(delay(0));
        },
        LastTrigger() {
            return of([]).pipe(delay(0));
        },
        ListWebhookPoliciesOfProjectResponse() {
            const response: HttpResponse<Array<Registry>> = new HttpResponse<
                Array<Registry>
            >({
                headers: new HttpHeaders({
                    'x-total-count': [mockedWehook].length.toString(),
                }),
                body: [mockedWehook],
            });
            return of(response).pipe(delay(0));
        },
        UpdateWebhookPolicyOfProject() {
            return of(true);
        },
    };
    const mockActivatedRoute = {
        RouterparamMap: of({ get: key => 'value' }),
        snapshot: {
            parent: {
                parent: {
                    params: { id: 1 },
                    data: {
                        projectResolver: {
                            ismember: true,
                            name: 'library',
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

    beforeEach(() => {
        TestBed.configureTestingModule({
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
            imports: [SharedTestingModule],
            declarations: [
                WebhookComponent,
                AddWebhookComponent,
                AddWebhookFormComponent,
                InlineAlertComponent,
                ConfirmationDialogComponent,
            ],
            providers: [
                {
                    provide: ProjectWebhookService,
                    useValue: mockProjectWebhookService,
                },
                { provide: WebhookService, useValue: mockedWebhookService },
                {
                    provide: MessageHandlerService,
                    useValue: mockMessageHandlerService,
                },
                { provide: ActivatedRoute, useValue: mockActivatedRoute },
                {
                    provide: UserPermissionService,
                    useValue: mockUserPermissionService,
                },
            ],
        }).compileComponents();
    });

    beforeEach(async () => {
        fixture = TestBed.createComponent(WebhookComponent);
        component = fixture.componentInstance;
        fixture.autoDetectChanges(true);
        await fixture.whenStable();
    });

    it('should create', async () => {
        expect(component).toBeTruthy();
    });
    it('should get webhook list', async () => {
        const rows = fixture.nativeElement.getElementsByTagName('clr-dg-row');
        expect(rows.length).toEqual(1);
    });
    it('should open modal', async () => {
        component.newWebhook();
        fixture.detectChanges();
        await fixture.whenStable();
        const body: HTMLElement =
            fixture.nativeElement.querySelector('.modal-body');
        expect(body).toBeTruthy();
        const title: HTMLElement =
            fixture.nativeElement.querySelector('.modal-title');
        expect(title.innerText).toEqual('WEBHOOK.ADD_WEBHOOK');
    });
    it('should open edit modal', async () => {
        component.webhookList[0].name = 'test';
        component.selectedRow[0] = component.webhookList[0];
        component.editWebhook();
        fixture.detectChanges();
        await fixture.whenStable();
        const body: HTMLElement =
            fixture.nativeElement.querySelector('.modal-body');
        expect(body).toBeTruthy();
        const title: HTMLElement =
            fixture.nativeElement.querySelector('.modal-title');
        expect(title.innerText).toEqual('WEBHOOK.EDIT_WEBHOOK');
        const nameInput: HTMLInputElement =
            fixture.nativeElement.querySelector('#name');
        expect(nameInput.value).toEqual('test');
    });
    it('should disable webhook', async () => {
        await fixture.whenStable();
        component.selectedRow[0] = component.webhookList[0];
        component.webhookList[0].enabled = true;
        component.switchWebhookStatus();
        fixture.detectChanges();
        await fixture.whenStable();
        const button: HTMLButtonElement = fixture.nativeElement.querySelector(
            '#dialog-action-disable'
        );
        button.dispatchEvent(new Event('click'));
        await fixture.whenStable();
        const body: HTMLElement =
            fixture.nativeElement.querySelector('.modal-body');
        expect(body).toBeFalsy();
    });
    it('should enable webhook', async () => {
        await fixture.whenStable();
        component.webhookList[0].enabled = false;
        component.selectedRow[0] = component.webhookList[0];
        component.switchWebhookStatus();
        fixture.detectChanges();
        await fixture.whenStable();
        const buttonEnable: HTMLButtonElement =
            fixture.nativeElement.querySelector('#dialog-action-enable');
        buttonEnable.dispatchEvent(new Event('click'));
        await fixture.whenStable();
        const bodyEnable: HTMLElement =
            fixture.nativeElement.querySelector('.modal-body');
        expect(bodyEnable).toBeFalsy();
    });
});
