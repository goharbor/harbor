import { ComponentFixture, TestBed } from '@angular/core/testing';
import { TranslateService } from '@ngx-translate/core';
import { ChartDetailSummaryComponent } from './chart-detail-summary.component';
import { CUSTOM_ELEMENTS_SCHEMA, SecurityContext } from '@angular/core';
import { MarkdownModule, MarkedOptions } from 'ngx-markdown';
import { HelmChartService } from '../helm-chart.service';
import { ErrorHandler } from '../../../../../shared/units/error-handler';
import { MessageHandlerService } from '../../../../../shared/services/message-handler.service';
import { SharedTestingModule } from '../../../../../shared/shared.module';

describe('ChartDetailSummaryComponent', () => {
    let component: ChartDetailSummaryComponent;
    let fixture: ComponentFixture<ChartDetailSummaryComponent>;
    const mockHelmChartService = {
        downloadChart: function () {},
    };

    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [
                SharedTestingModule,
                MarkdownModule.forRoot({ sanitize: SecurityContext.HTML }),
            ],
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
            declarations: [ChartDetailSummaryComponent],
            providers: [
                TranslateService,
                { provide: MarkedOptions, useValue: {} },
                { provide: ErrorHandler, useValue: MessageHandlerService },
                { provide: HelmChartService, useValue: mockHelmChartService },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ChartDetailSummaryComponent);
        component = fixture.componentInstance;
        component.summary = {
            name: 'string',
            home: 'string',
            sources: [],
            version: 'string',
            description: 'string',
            keywords: [],
            maintainers: [],
            engine: 'string',
            icon: 'string',
            appVersion: 'string',
            urls: [],
            created: new Date().toDateString(),
            digest: 'string',
        };
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
