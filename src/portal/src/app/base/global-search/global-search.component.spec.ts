import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { GlobalSearchComponent } from './global-search.component';

xdescribe('GlobalSearchComponent', () => {
    let component: GlobalSearchComponent;
    let fixture: ComponentFixture<GlobalSearchComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [GlobalSearchComponent]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(GlobalSearchComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
