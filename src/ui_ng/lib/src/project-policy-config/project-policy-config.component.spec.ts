import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { ErrorHandler } from '../error-handler/error-handler';
import { ConfirmationDialogComponent } from '../confirmation-dialog/confirmation-dialog.component';
import { ProjectPolicyConfigComponent } from './project-policy-config.component';
import { SharedModule } from '../shared/shared.module';
import { ProjectService, ProjectDefaultService} from '../service/project.service';
import { SERVICE_CONFIG, IServiceConfig} from '../service.config';

describe('ProjectPolicyConfigComponent', () => {
  let component: ProjectPolicyConfigComponent;
  let fixture: ComponentFixture<ProjectPolicyConfigComponent>;

  let config: IServiceConfig = {
    projectPolicyEndpoint: '/api/projects/testing/:id/'
  };

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [SharedModule],
      declarations: [
        ProjectPolicyConfigComponent,
        ConfirmationDialogComponent,
        ConfirmationDialogComponent,
       ],
       providers: [
        ErrorHandler,
        { provide: SERVICE_CONFIG, useValue: config },
        { provide: ProjectService, useClass: ProjectDefaultService }
      ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ProjectPolicyConfigComponent);
    component = fixture.componentInstance;
    component.projectId = 1;
    component.hasProjectAdminRole = true;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
