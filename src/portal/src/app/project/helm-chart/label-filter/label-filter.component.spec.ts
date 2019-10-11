import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { LabelFilterComponent } from './label-filter.component';

xdescribe('LabelFilterComponent', () => {
    let component: LabelFilterComponent;
    let fixture: ComponentFixture<LabelFilterComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [LabelFilterComponent]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(LabelFilterComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
