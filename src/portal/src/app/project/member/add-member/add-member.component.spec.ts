import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { AddMemberComponent } from './add-member.component';

xdescribe('AddMemberComponent', () => {
    let component: AddMemberComponent;
    let fixture: ComponentFixture<AddMemberComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [AddMemberComponent]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(AddMemberComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
