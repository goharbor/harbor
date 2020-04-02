import { ComponentFixture, TestBed } from '@angular/core/testing';
import { BrowserAnimationsModule } from "@angular/platform-browser/animations";
import { ClarityModule } from "@clr/angular";
import { SharedModule } from "../../../shared/shared.module";
import { LastTriggerComponent } from "./last-trigger.component";
import { LastTrigger } from "../webhook";
import { SimpleChange } from "@angular/core";
import { of } from "rxjs";
import { WebhookService } from "../webhook.service";

describe('LastTriggerComponent', () => {
  const mokedTriggers: LastTrigger[] = [
    {
      policy_name: 'http',
      enabled: true,
      event_type: 'pullImage',
      creation_time: null,
      last_trigger_time: null
    },
    {
      policy_name: 'slack',
      enabled: true,
      event_type: 'pullImage',
      creation_time: null,
      last_trigger_time: null
    }
  ];
  const mockWebhookService = {
    eventTypeToText(eventType: string) {
      return eventType;
    }
  };
  let component: LastTriggerComponent;
  let fixture: ComponentFixture<LastTriggerComponent>;
  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [
        SharedModule,
        BrowserAnimationsModule,
        ClarityModule,
      ],
      declarations: [
        LastTriggerComponent
      ],
      providers: [{ provide: WebhookService, useValue: mockWebhookService }]
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
    component.ngOnChanges({inputLastTriggers: new SimpleChange([], mokedTriggers, true)});
    fixture.detectChanges();
    await fixture.whenStable();
    const rows = fixture.nativeElement.getElementsByTagName('clr-dg-row');
    expect(rows.length).toEqual(1);
  });
});
