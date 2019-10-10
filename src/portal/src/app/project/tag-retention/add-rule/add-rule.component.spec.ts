import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { AddRuleComponent } from './add-rule.component';

xdescribe('AddRuleComponent', () => {
    let component: AddRuleComponent;
    let fixture: ComponentFixture<AddRuleComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [AddRuleComponent]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(AddRuleComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
