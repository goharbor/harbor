import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { ValuesComponent } from './chart-detail-value.component';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { ClarityModule } from '@clr/angular';
import { FormsModule } from '@angular/forms';
import { MarkdownModule, MarkdownService, MarkedOptions  } from 'ngx-markdown';
import { BrowserModule } from '@angular/platform-browser';

describe('ChartDetailValueComponent', () => {
    let component: ValuesComponent;
    let fixture: ComponentFixture<ValuesComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [
                TranslateModule.forRoot(),
                MarkdownModule,
                ClarityModule,
                FormsModule,
                BrowserModule
            ],
            declarations: [ValuesComponent],
            schemas: [
                CUSTOM_ELEMENTS_SCHEMA
            ],
            providers: [
                TranslateService,
                MarkdownService,
                { provide: MarkedOptions, useValue: {} },
            ]
        }).compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(ValuesComponent);
        component = fixture.componentInstance;
        component.yaml = "rfrf";
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
