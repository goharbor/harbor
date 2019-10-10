import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { RepositoryPageComponent } from './repository-page.component';

xdescribe('RepositoryPageComponent', () => {
    let component: RepositoryPageComponent;
    let fixture: ComponentFixture<RepositoryPageComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [RepositoryPageComponent]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(RepositoryPageComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
