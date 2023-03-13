import { ComponentFixture, TestBed } from '@angular/core/testing';
import { NO_ERRORS_SCHEMA } from '@angular/core';
import { DonutChartComponent } from './donut-chart.component';

describe('DonutChartComponent', () => {
    let component: DonutChartComponent;
    let fixture: ComponentFixture<DonutChartComponent>;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            schemas: [NO_ERRORS_SCHEMA],
            declarations: [DonutChartComponent],
        }).compileComponents();

        fixture = TestBed.createComponent(DonutChartComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
