import { async, ComponentFixture, TestBed, ComponentFixtureAutoDetect } from '@angular/core/testing';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { ListProjectComponent } from './list-project.component';
import { CUSTOM_ELEMENTS_SCHEMA, ChangeDetectorRef } from '@angular/core';
import { ClarityModule } from '@clr/angular';
import { FormsModule } from '@angular/forms';
import { OperationService, ProjectService } from "@harbor/ui";
import { SessionService } from "../../shared/session.service";
import { AppConfigService } from "../../app-config.service";
import { RouterTestingModule } from '@angular/router/testing';
import { SearchTriggerService } from "../../base/global-search/search-trigger.service";
import { MessageHandlerService } from "../../shared/message-handler/message-handler.service";
import { StatisticHandler } from "../../shared/statictics/statistic-handler.service";
import { ConfirmationDialogService } from "../../shared/confirmation-dialog/confirmation-dialog.service";
import { of } from 'rxjs';
import { BrowserAnimationsModule, NoopAnimationsModule } from "@angular/platform-browser/animations";
import { delay } from 'rxjs/operators';
describe('ListProjectComponent', () => {
    let component: ListProjectComponent;
    let fixture: ComponentFixture<ListProjectComponent>;
    const mockProjectService = {
        listProjects: () => {
            return of({
                body: []
            }).pipe(delay(0));
        }
    };
    const mockSessionService = {
        getCurrentUser: () => {
            return false;
        }
    };
    const mockAppConfigService = {
        getConfig: () => {
            return {
                project_creation_restriction: "",
                with_chartmuseum: ""
            };
        }
    };
    const mockSearchTriggerService = {
        closeSearch: () => {
        }
    };
    const mockStatisticHandler = {
        handleError: () => {
        }
    };
    const mockMessageHandlerService = {
        refresh: () => {
        },
        showSuccess: () => {
        },
    };
    const mockConfirmationDialogService = {
        confirmationConfirm$: of({
            state: "",
            data: []
        })
    };
    const mockOperationService = {
        publishInfo$: () => {}
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
                NoopAnimationsModule
            ],
            declarations: [ListProjectComponent],
            providers: [
                TranslateService,
                ChangeDetectorRef,
                { provide: ProjectService, useValue: mockProjectService },
                { provide: SessionService, useValue: mockSessionService },
                { provide: AppConfigService, useValue: mockAppConfigService },
                { provide: SearchTriggerService, useValue: mockSearchTriggerService },
                { provide: MessageHandlerService, useValue: mockMessageHandlerService },
                { provide: StatisticHandler, useValue: mockStatisticHandler },
                { provide: ConfirmationDialogService, useValue: mockConfirmationDialogService },
                { provide: OperationService, useValue: mockOperationService },
                { provide: ComponentFixtureAutoDetect, useValue: true }

            ]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(ListProjectComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });
    let originalTimeout;
    beforeEach(function () {
        originalTimeout = jasmine.DEFAULT_TIMEOUT_INTERVAL;
        jasmine.DEFAULT_TIMEOUT_INTERVAL = 100000;
      });

      afterEach(function () {
        jasmine.DEFAULT_TIMEOUT_INTERVAL = originalTimeout;
      });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
