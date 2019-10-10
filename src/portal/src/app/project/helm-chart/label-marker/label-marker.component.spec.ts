import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { LabelMarkerComponent } from './label-marker.component';

xdescribe('LabelMarkerComponent', () => {
    let component: LabelMarkerComponent;
    let fixture: ComponentFixture<LabelMarkerComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [LabelMarkerComponent]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(LabelMarkerComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
