import { ComponentFixture, TestBed } from '@angular/core/testing';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { ChartDetailValueComponent } from './chart-detail-value.component';
import { CUSTOM_ELEMENTS_SCHEMA, SecurityContext } from '@angular/core';
import { ClarityModule } from '@clr/angular';
import { FormsModule } from '@angular/forms';
import { MarkdownModule, MarkedOptions } from 'ngx-markdown';
import { BrowserModule } from '@angular/platform-browser';

describe('ChartDetailValueComponent', () => {
    let component: ChartDetailValueComponent;
    let fixture: ComponentFixture<ChartDetailValueComponent>;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [
                TranslateModule.forRoot(),
                ClarityModule,
                FormsModule,
                BrowserModule,
                MarkdownModule.forRoot({ sanitize: SecurityContext.HTML }),
            ],
            declarations: [ChartDetailValueComponent],
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
            providers: [
                TranslateService,
                { provide: MarkedOptions, useValue: {} },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ChartDetailValueComponent);
        component = fixture.componentInstance;
        component.yaml = 'rfrf';
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
