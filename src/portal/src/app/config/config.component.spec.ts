import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { SessionService } from '../shared/session.service';
import { ConfirmationDialogService } from '../shared/confirmation-dialog/confirmation-dialog.service';
import { MessageHandlerService } from '../shared/message-handler/message-handler.service';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { ClarityModule } from "@clr/angular";
import { AppConfigService } from '../app-config.service';
import { ConfigurationService } from './config.service';
import { ConfigurationComponent } from './config.component';

describe('ConfigurationComponent', () => {
    let component: ConfigurationComponent;
    let fixture: ComponentFixture<ConfigurationComponent>;
    let fakeConfirmationDialogService = {
        confirmationConfirm$: {
            subscribe: function () {
            }
        }
    };

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [
                TranslateModule.forRoot(),
                ClarityModule
            ],
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
            declarations: [ConfigurationComponent],
            providers: [
                TranslateService,
                {
                    provide: SessionService, useValue: {
                        getCurrentUser: function () {
                            return "admin";
                        }
                    }
                },
                { provide: ConfirmationDialogService, useValue: fakeConfirmationDialogService },
                { provide: MessageHandlerService, useValue: null },
                {
                    provide: AppConfigService, useValue: {
                        getConfig: function () {
                            return { has_ca_root: true };
                        }
                    }
                },
                {
                    provide: ConfigurationService, useValue: {
                        confirmationConfirm$: {
                            subscribe: function () {
                            }
                        }
                    }
                }
            ]
        }).compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(ConfigurationComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
