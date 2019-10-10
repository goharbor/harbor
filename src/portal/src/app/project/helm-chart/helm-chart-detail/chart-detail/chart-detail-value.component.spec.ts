import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { ChartDetailValueComponent } from './chart-detail-value.component';

xdescribe('ChartDetailValueComponent', () => {
    let component: ChartDetailValueComponent;
    let fixture: ComponentFixture<ChartDetailValueComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [ChartDetailValueComponent]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(ChartDetailValueComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
