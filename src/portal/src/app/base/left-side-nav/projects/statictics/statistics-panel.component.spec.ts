import { ComponentFixture, TestBed } from '@angular/core/testing';
import { StatisticsPanelComponent } from './statistics-panel.component';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { of } from 'rxjs';
import { SessionService } from '../../../../shared/services/session.service';
import { MessageHandlerService } from '../../../../shared/services/message-handler.service';
import { StatisticHandler } from './statistic-handler.service';
import { AppConfigService } from '../../../../services/app-config.service';
import { Statistic } from '../../../../../../ng-swagger-gen/models/statistic';
import { SharedTestingModule } from '../../../../shared/shared.module';
import { StatisticService } from '../../../../../../ng-swagger-gen/services/statistic.service';

describe('StatisticsPanelComponent', () => {
    const mockedStatistic: Statistic = {
        private_project_count: 2,
        private_repo_count: 0,
        public_project_count: 3,
        public_repo_count: 1,
        total_project_count: 5,
        total_repo_count: 1,
        total_storage_consumption: 4564,
    };
    let component: StatisticsPanelComponent;
    let fixture: ComponentFixture<StatisticsPanelComponent>;
    const mockStatisticsService = {
        getStatistic: () => of(mockedStatistic),
    };
    const mockSessionService = {
        getCurrentUser: () => {
            return {
                has_admin_role: true,
            };
        },
    };
    const mockAppConfigService = {
        getConfig: () => {
            return {
                registry_storage_provider_name: '',
            };
        },
    };
    const mockMessageHandlerService = {
        handleError: () => {},
    };
    const mockStatisticHandler = {
        refreshChan$: of(null),
    };
    beforeEach(() => {
        TestBed.configureTestingModule({
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
            imports: [SharedTestingModule],
            declarations: [StatisticsPanelComponent],
            providers: [
                { provide: SessionService, useValue: mockSessionService },
                { provide: AppConfigService, useValue: mockAppConfigService },
                { provide: StatisticService, useValue: mockStatisticsService },
                { provide: StatisticHandler, useValue: mockStatisticHandler },
                {
                    provide: MessageHandlerService,
                    useValue: mockMessageHandlerService,
                },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(StatisticsPanelComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
    it('should have 3 cards', async () => {
        fixture.detectChanges();
        await fixture.whenStable();
        const cards = fixture.nativeElement.querySelectorAll('.card');
        expect(cards.length).toEqual(3);
    });
    it('should display right size number', async () => {
        fixture.detectChanges();
        await fixture.whenStable();
        const sizeHtml: HTMLSpanElement =
            fixture.nativeElement.querySelector('.size-number');
        expect(sizeHtml.innerText).toEqual('4.46');
    });
});
