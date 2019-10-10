import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { ListChartVersionRoComponent } from './list-chart-version-ro.component';

xdescribe('ListChartVersionRoComponent', () => {
    let component: ListChartVersionRoComponent;
    let fixture: ComponentFixture<ListChartVersionRoComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [ListChartVersionRoComponent]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(ListChartVersionRoComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
