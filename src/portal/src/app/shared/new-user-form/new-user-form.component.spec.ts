import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { NewUserFormComponent } from './new-user-form.component';

xdescribe('NewUserFormComponent', () => {
    let component: NewUserFormComponent;
    let fixture: ComponentFixture<NewUserFormComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [NewUserFormComponent]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(NewUserFormComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
