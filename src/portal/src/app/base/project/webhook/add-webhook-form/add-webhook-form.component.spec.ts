import { ComponentFixture, TestBed } from '@angular/core/testing';
import { AddWebhookFormComponent } from './add-webhook-form.component';
import { ProjectWebhookService } from '../webhook.service';
import { MessageHandlerService } from '../../../../shared/services/message-handler.service';
import { of } from 'rxjs';
import { InlineAlertComponent } from '../../../../shared/components/inline-alert/inline-alert.component';
import { WebhookPolicy } from '../../../../../../ng-swagger-gen/models/webhook-policy';
import { delay } from 'rxjs/operators';
import { HttpHeaders, HttpResponse } from '@angular/common/http';
import { Registry } from '../../../../../../ng-swagger-gen/models/registry';
import { SharedTestingModule } from '../../../../shared/shared.module';
import { WebhookService } from '../../../../../../ng-swagger-gen/services/webhook.service';

describe('AddWebhookFormComponent', () => {
    let component: AddWebhookFormComponent;
    let fixture: ComponentFixture<AddWebhookFormComponent>;
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
        ListWebhookPoliciesOfProject: () => {
            return of(null);
        },
        UpdateWebhookPolicyOfProject() {
            return of(true);
        },
        CreateWebhookPolicyOfProject() {
            return of(true);
        },
    };
    const mockMessageHandlerService = {
        handleError: () => {},
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

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [AddWebhookFormComponent, InlineAlertComponent],
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
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(AddWebhookFormComponent);
        component = fixture.componentInstance;
        component.metadata = mockedMetadata;
        fixture.detectChanges();
    });

    it('should create', async () => {
        expect(component).toBeTruthy();
        const cancelButtonForAdd: HTMLButtonElement =
            fixture.nativeElement.querySelector('#add-webhook-cancel');
        expect(cancelButtonForAdd).toBeTruthy();
        component.isModify = true;
        fixture.detectChanges();
        await fixture.whenStable();
        const cancelButtonForEdit: HTMLButtonElement =
            fixture.nativeElement.querySelector('#edit-webhook-cancel');
        expect(cancelButtonForEdit).toBeTruthy();
    });
    it("should occur a 'name is required' error", async () => {
        await fixture.whenStable();
        fixture.autoDetectChanges(true);
        const nameInput: HTMLInputElement =
            fixture.nativeElement.querySelector('#name');
        nameInput.value = 'test';
        nameInput.dispatchEvent(new Event('input'));
        nameInput.value = null;
        nameInput.dispatchEvent(new Event('input'));
        nameInput.blur();
        nameInput.dispatchEvent(new Event('blur'));
        const errorEle: HTMLElement =
            fixture.nativeElement.querySelector('clr-control-error');
        expect(errorEle.innerText).toEqual('WEBHOOK.NAME_REQUIRED');
    });
    it('add button should work', async () => {
        const spy: jasmine.Spy = spyOn(component, 'add').and.returnValue(
            undefined
        );
        component.webhook = mockedWehook;
        fixture.detectChanges();
        await fixture.whenStable();
        expect(component.isValid).toBeTruthy();
        const addButton: HTMLButtonElement =
            fixture.nativeElement.querySelector('#new-webhook-continue');
        addButton.dispatchEvent(new Event('click'));
        fixture.detectChanges();
        await fixture.whenStable();
        expect(spy.calls.count()).toEqual(1);
    });
});
