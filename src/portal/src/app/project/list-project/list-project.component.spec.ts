import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { ListProjectComponent } from './list-project.component';

xdescribe('ListProjectComponent', () => {
    let component: ListProjectComponent;
    let fixture: ComponentFixture<ListProjectComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [ListProjectComponent]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(ListProjectComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
