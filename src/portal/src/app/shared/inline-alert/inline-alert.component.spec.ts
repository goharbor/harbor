import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { InlineAlertComponent } from './inline-alert.component';

xdescribe('InlineAlertComponent', () => {
    let component: InlineAlertComponent;
    let fixture: ComponentFixture<InlineAlertComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [InlineAlertComponent]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(InlineAlertComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
