import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { TopRepoComponent } from './top-repo.component';

xdescribe('TopRepoComponent', () => {
    let component: TopRepoComponent;
    let fixture: ComponentFixture<TopRepoComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [TopRepoComponent]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(TopRepoComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
