import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { WebhookComponent } from './webhook.component';
import { ActivatedRoute } from '@angular/router';
import { WebhookService } from './webhook.service';
import { MessageHandlerService } from "../../shared/message-handler/message-handler.service";
import { of } from 'rxjs';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { BrowserAnimationsModule, NoopAnimationsModule } from '@angular/platform-browser/animations';
import { ClarityModule } from '@clr/angular';
import { FormsModule } from '@angular/forms';
import { RouterTestingModule } from '@angular/router/testing';
import { HttpClientTestingModule } from '@angular/common/http/testing';
describe('WebhookComponent', () => {
    let component: WebhookComponent;
    let fixture: ComponentFixture<WebhookComponent>;
    const mockMessageHandlerService = {
        handleError: () => { }
    };
    const mockWebhookService = {
        listLastTrigger: () => {
            return of([]);
        },
        listWebhook: () => {
            return of([
                {
                    targets: [
                        { address: "" }
                    ],
                    enabled: true
                }
            ]);
        },
    };
    const mockActivatedRoute = {
        RouterparamMap: of({ get: (key) => 'value' }),
        snapshot: {
            parent: {
                params: { id: 1 },
                data: {
                    projectResolver: {
                        ismember: true,
                        name: 'library',
                    }
                }
            }
        }
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
            declarations: [WebhookComponent],
            providers: [
                TranslateService,
                { provide: WebhookService, useValue: mockWebhookService },
                { provide: MessageHandlerService, useValue: mockMessageHandlerService },
                { provide: ActivatedRoute, useValue: mockActivatedRoute },
            ]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(WebhookComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
