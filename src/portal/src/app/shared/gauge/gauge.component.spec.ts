import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { GaugeComponent } from './gauge.component';

xdescribe('GaugeComponent', () => {
    let component: GaugeComponent;
    let fixture: ComponentFixture<GaugeComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [GaugeComponent]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(GaugeComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
