import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { StatisticsPanelComponent } from './statistics-panel.component';
import { StatisticsComponent } from './statistics.component';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { BrowserAnimationsModule, NoopAnimationsModule } from '@angular/platform-browser/animations';
import { ClarityModule } from '@clr/angular';
import { FormsModule } from '@angular/forms';
import { RouterTestingModule } from '@angular/router/testing';
import { of } from 'rxjs';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { StatisticsService } from "./statistics.service";
import { SessionService } from "../session.service";
import { MessageHandlerService } from "../message-handler/message-handler.service";
import { StatisticHandler } from "./statistic-handler.service";
import { AppConfigService } from "./../../app-config.service";
import { Statistics } from './statistics';
import { Volumes } from './volumes';
describe('StatisticsPanelComponent', () => {
    let component: StatisticsPanelComponent;
    let fixture: ComponentFixture<StatisticsPanelComponent>;
    const mockStatisticsService = {
        getStatistics: () => of(new Statistics()),
        getVolumes: () => of(new Volumes()),
    };
    const mockSessionService = {
        getCurrentUser: () => { }
    };
    const mockAppConfigService = {
        getConfig: () => {
            return {
                registry_storage_provider_name : ""
            };
        }
    };
    const mockMessageHandlerService = {
        handleError: () => { }
    };
    const mockStatisticHandler = {
        refreshChan$: of(null)
    };
    const mockRouter = {
        navigate: () => { }
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
            declarations: [StatisticsPanelComponent, StatisticsComponent],
            providers: [
                TranslateService,
                { provide: SessionService, useValue: mockSessionService },
                { provide: AppConfigService, useValue: mockAppConfigService },
                { provide: StatisticsService, useValue: mockStatisticsService },
                { provide: StatisticHandler, useValue: mockStatisticHandler },
                { provide: MessageHandlerService, useValue: mockMessageHandlerService }
            ]
        }).compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(StatisticsPanelComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
