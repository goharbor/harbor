import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { ChartDetailDependencyComponent } from './chart-detail-dependency.component';

xdescribe('ChartDetailDependencyComponent', () => {
    let component: ChartDetailDependencyComponent;
    let fixture: ComponentFixture<ChartDetailDependencyComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [ChartDetailDependencyComponent]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(ChartDetailDependencyComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
