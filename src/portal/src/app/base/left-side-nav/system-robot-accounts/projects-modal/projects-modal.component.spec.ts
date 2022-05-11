import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ProjectsModalComponent } from './projects-modal.component';
import { Project } from '../../../../../../ng-swagger-gen/models/project';
import { Action, PermissionsKinds, Resource } from '../system-robot-util';
import { RobotPermission } from '../../../../../../ng-swagger-gen/models/robot-permission';
import { SharedTestingModule } from '../../../../shared/shared.module';

describe('ProjectsModalComponent', () => {
    let component: ProjectsModalComponent;
    let fixture: ComponentFixture<ProjectsModalComponent>;
    const project1: Project = {
        project_id: 1,
        name: 'project1',
    };
    const project2: Project = {
        project_id: 2,
        name: 'project2',
    };
    const permissions: RobotPermission[] = [
        {
            kind: PermissionsKinds.PROJECT,
            namespace: project1.name,
            access: [
                {
                    resource: Resource.ARTIFACT,
                    action: Action.PUSH,
                },
            ],
        },
        {
            kind: PermissionsKinds.PROJECT,
            namespace: project2.name,
            access: [
                {
                    resource: Resource.ARTIFACT,
                    action: Action.PUSH,
                },
            ],
        },
    ];
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [ProjectsModalComponent],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ProjectsModalComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should render list', async () => {
        component.projectsModalOpened = true;
        component.permissions = permissions;
        fixture.detectChanges();
        await fixture.whenStable();
        const rows = fixture.nativeElement.querySelectorAll('clr-dg-row');
        expect(rows.length).toEqual(2);
    });
});
