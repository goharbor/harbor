import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { ListProjectROComponent } from './list-project-ro.component';

xdescribe('ListProjectROComponent', () => {
    let component: ListProjectROComponent;
    let fixture: ComponentFixture<ListProjectROComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [ListProjectROComponent]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(ListProjectROComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
