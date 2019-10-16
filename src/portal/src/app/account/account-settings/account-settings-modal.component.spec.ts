import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { RouterTestingModule } from '@angular/router/testing';
import { AccountSettingsModalComponent } from './account-settings-modal.component';
import { SessionService } from "../../shared/session.service";
import { MessageHandlerService } from "../../shared/message-handler/message-handler.service";
import { SearchTriggerService } from "../../base/global-search/search-trigger.service";
import { AccountSettingsModalService } from './account-settings-modal-service.service';
import { CUSTOM_ELEMENTS_SCHEMA, ChangeDetectorRef } from '@angular/core';
import { ClarityModule } from "@clr/angular";
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { FormsModule } from '@angular/forms';

describe('AccountSettingsModalComponent', () => {
    let component: AccountSettingsModalComponent;
    let fixture: ComponentFixture<AccountSettingsModalComponent>;
    let fakeSessionService = {
        getCurrentUser: function () {
            return { has_admin_role: true };
        }
    };
    let fakeMessageHandlerService = null;
    let fakeSearchTriggerService = null;
    let fakeAccountSettingsModalService = null;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [AccountSettingsModalComponent],
            imports: [
                RouterTestingModule,
                ClarityModule,
                TranslateModule.forRoot(),
                FormsModule
            ],
            providers: [
                ChangeDetectorRef,
                TranslateService,
                { provide: SessionService, useValue: fakeSessionService },
                { provide: MessageHandlerService, useValue: fakeMessageHandlerService },
                { provide: SearchTriggerService, useValue: fakeSearchTriggerService },
                { provide: AccountSettingsModalService, useValue: fakeAccountSettingsModalService }
            ],
            schemas: [
                CUSTOM_ELEMENTS_SCHEMA
            ],
        }).compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(AccountSettingsModalComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
