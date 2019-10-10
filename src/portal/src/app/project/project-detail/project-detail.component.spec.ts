import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { ProjectDetailComponent } from './project-detail.component';

xdescribe('ProjectDetailComponent', () => {
    let component: ProjectDetailComponent;
    let fixture: ComponentFixture<ProjectDetailComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [ProjectDetailComponent]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(ProjectDetailComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
