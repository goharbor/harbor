import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { AddWebhookFormComponent } from './add-webhook-form.component';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { BrowserAnimationsModule, NoopAnimationsModule } from '@angular/platform-browser/animations';
import { RouterTestingModule } from '@angular/router/testing';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { ClarityModule } from '@clr/angular';
import { FormsModule } from '@angular/forms';
import { WebhookService } from "../webhook.service";
import { MessageHandlerService } from "../../../shared/message-handler/message-handler.service";
import { of } from 'rxjs';

describe('AddWebhookFormComponent', () => {
    let component: AddWebhookFormComponent;
    let fixture: ComponentFixture<AddWebhookFormComponent>;
    const mockWebhookService = {
        getCurrentUser: () => {
            return of(null);
        }
    };
    const mockMessageHandlerService = {
        handleError: () => { }
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
            declarations: [AddWebhookFormComponent],
            providers: [
                TranslateService,
                { provide: WebhookService, useValue: mockWebhookService },
                { provide: MessageHandlerService, useValue: mockMessageHandlerService },


            ]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(AddWebhookFormComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
