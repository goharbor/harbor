import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { ChartDetailSummaryComponent } from './chart-detail-summary.component';
import { ClarityModule } from '@clr/angular';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { MarkedOptions, MarkdownModule, MarkdownService } from 'ngx-markdown';
import { ErrorHandler, DefaultErrorHandler } from '@harbor/ui';
import { HelmChartService } from '../../helm-chart.service';
describe('ChartDetailSummaryComponent', () => {
    let component: ChartDetailSummaryComponent;
    let fixture: ComponentFixture<ChartDetailSummaryComponent>;
    const mockHelmChartService = {
        downloadChart: function () {
        }
    };

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [
                TranslateModule.forRoot(),
                ClarityModule,
                FormsModule,
                MarkdownModule,
            ],
            schemas: [
                CUSTOM_ELEMENTS_SCHEMA
            ],
            declarations: [ChartDetailSummaryComponent],
            providers: [
                TranslateService,
                MarkdownService,
                { provide: MarkedOptions, useValue: {} },
                { provide: ErrorHandler, useValue: DefaultErrorHandler },
                { provide: HelmChartService, useValue: mockHelmChartService },
            ]
        }).compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(ChartDetailSummaryComponent);
        // markdownService = TestBed.get(MarkdownService);
        component = fixture.componentInstance;
        component.summary = {
            name: "string",
            home: "string",
            sources: [],
            version: "string",
            description: "string",
            keywords: [],
            maintainers: [],
            engine: "string",
            icon: "string",
            appVersion: "string",
            urls: [],
            created: new Date().toDateString(),
            digest: "string",
        };
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
