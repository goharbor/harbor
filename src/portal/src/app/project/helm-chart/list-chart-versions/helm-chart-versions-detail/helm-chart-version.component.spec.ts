import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { ChartVersionComponent } from './helm-chart-version.component';

xdescribe('ChartVersionComponent', () => {
    let component: ChartVersionComponent;
    let fixture: ComponentFixture<ChartVersionComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [ChartVersionComponent]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(ChartVersionComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
