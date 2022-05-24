import { ComponentFixture, TestBed } from '@angular/core/testing';
import { SharedTestingModule } from '../../../../shared/shared.module';
import { LastTriggerComponent } from './last-trigger.component';
import { SimpleChange } from '@angular/core';
import { ProjectWebhookService } from '../webhook.service';
import { WebhookLastTrigger } from '../../../../../../ng-swagger-gen/models/webhook-last-trigger';

describe('LastTriggerComponent', () => {
    const mokedTriggers: WebhookLastTrigger[] = [
        {
            policy_name: 'http',
            enabled: true,
            event_type: 'pullImage',
            creation_time: null,
            last_trigger_time: null,
        },
        {
            policy_name: 'slack',
            enabled: true,
            event_type: 'pullImage',
            creation_time: null,
            last_trigger_time: null,
        },
    ];
    const mockWebhookService = {
        eventTypeToText(eventType: string) {
            return eventType;
        },
    };
    let component: LastTriggerComponent;
    let fixture: ComponentFixture<LastTriggerComponent>;
    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [LastTriggerComponent],
            providers: [
                {
                    provide: ProjectWebhookService,
                    useValue: mockWebhookService,
                },
            ],
        });
    });
    beforeEach(() => {
        fixture = TestBed.createComponent(LastTriggerComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });
    it('should create', () => {
        expect(component).toBeTruthy();
    });
    it('should render one row', async () => {
        component.inputLastTriggers = mokedTriggers;
        component.webhookName = 'slack';
        component.ngOnChanges({
            inputLastTriggers: new SimpleChange([], mokedTriggers, true),
        });
        fixture.detectChanges();
        await fixture.whenStable();
        const rows = fixture.nativeElement.getElementsByTagName('clr-dg-row');
        expect(rows.length).toEqual(1);
    });
});
