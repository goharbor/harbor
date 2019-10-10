import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { MemberComponent } from './member.component';

xdescribe('MemberComponent', () => {
    let component: MemberComponent;
    let fixture: ComponentFixture<MemberComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [MemberComponent]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(MemberComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
