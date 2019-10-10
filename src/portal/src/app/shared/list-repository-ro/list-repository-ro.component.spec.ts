import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { ListRepositoryROComponent } from './list-repository-ro.component';

xdescribe('ListRepositoryRoComponent', () => {
    let component: ListRepositoryROComponent;
    let fixture: ComponentFixture<ListRepositoryROComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [ListRepositoryROComponent]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(ListRepositoryROComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
