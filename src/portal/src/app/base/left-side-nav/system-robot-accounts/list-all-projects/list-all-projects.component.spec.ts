import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ListAllProjectsComponent } from './list-all-projects.component';
import { clone } from '../../../../shared/units/utils';
import { INITIAL_ACCESSES } from '../system-robot-util';
import { Project } from '../../../../../../ng-swagger-gen/models/project';
import { SharedTestingModule } from '../../../../shared/shared.module';

describe('ListAllProjectsComponent', () => {
    let component: ListAllProjectsComponent;
    let fixture: ComponentFixture<ListAllProjectsComponent>;
    const project1: Project = {
        project_id: 1,
        name: 'project1',
    };
    const project2: Project = {
        project_id: 2,
        name: 'project2',
    };
    const project3: Project = {
        project_id: 3,
        name: 'project3',
    };
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [ListAllProjectsComponent],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ListAllProjectsComponent);
        component = fixture.componentInstance;
        component.defaultAccesses = clone(INITIAL_ACCESSES);
        component.cachedAllProjects = [project1, project2, project3];
        component.init(false);
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
    it('should render list', async () => {
        fixture.detectChanges();
        await fixture.whenStable();
        const rows = fixture.nativeElement.querySelectorAll('clr-dg-row');
        expect(rows.length).toEqual(3);
    });
});
