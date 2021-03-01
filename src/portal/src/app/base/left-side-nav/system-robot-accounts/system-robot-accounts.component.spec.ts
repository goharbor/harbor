import { waitForAsync, ComponentFixture, TestBed } from '@angular/core/testing';
import {  NO_ERRORS_SCHEMA } from "@angular/core";
import { SystemRobotAccountsComponent } from './system-robot-accounts.component';
import { RobotService } from "../../../../../ng-swagger-gen/services/robot.service";
import { HttpHeaders, HttpResponse } from "@angular/common/http";
import { of, Subscription } from "rxjs";
import { delay } from "rxjs/operators";
import { Robot } from "../../../../../ng-swagger-gen/models/robot";
import { Action, PermissionsKinds, Resource } from "./system-robot-util";
import { Project } from "../../../../../ng-swagger-gen/models/project";
import { ProjectService } from "../../../../../ng-swagger-gen/services/project.service";
import { MessageHandlerService } from "../../../shared/services/message-handler.service";
import { OperationService } from "../../../shared/components/operation/operation.service";
import { ConfirmationDialogService } from "../../global-confirmation-dialog/confirmation-dialog.service";
import { SharedTestingModule } from "../../../shared/shared.module";

describe('SystemRobotAccountsComponent', () => {
  let component: SystemRobotAccountsComponent;
  let fixture: ComponentFixture<SystemRobotAccountsComponent>;
  const project1: Project = {
    project_id: 1,
    name: 'project1'
  };
  const project2: Project = {
    project_id: 2,
    name: 'project2'
  };
  const project3: Project = {
    project_id: 3,
    name: 'project3'
  };
  const robot1: Robot = {
    id: 1,
    name: 'robot1',
    level: PermissionsKinds.SYSTEM,
    disable: false,
    expires_at: (new Date().getTime() + 100000) % 1000,
    description: 'for test',
    secret: 'tthf54hfth4545dfgd5g454grd54gd54g',
    permissions: [
      {
        kind: PermissionsKinds.PROJECT,
        namespace: 'project1',
        access: [
          {
            resource: Resource.ARTIFACT,
            action: Action.PUSH
          }
        ]
      }
    ]
  };
  const robot2: Robot = {
    id: 2,
    name: 'robot2',
    level: PermissionsKinds.SYSTEM,
    disable: false,
    expires_at: (new Date().getTime() + 100000) % 1000,
    description: 'for test',
    secret: 'fsdf454654654fs6dfe',
    permissions: [
      {
        kind: PermissionsKinds.PROJECT,
        namespace: 'project2',
        access: [
          {
            resource: Resource.ARTIFACT,
            action: Action.PUSH
          }
        ]
      }
    ]
  };
  const robot3: Robot = {
    id: 3,
    name: 'robot3',
    level: PermissionsKinds.SYSTEM,
    disable: false,
    expires_at: (new Date().getTime() + 100000) % 1000,
    description: 'for test',
    secret: 'fsdg48454fse84',
    permissions: [
      {
        kind: PermissionsKinds.PROJECT,
        namespace: 'project3',
        access: [
          {
            resource: Resource.ARTIFACT,
            action: Action.PUSH
          }
        ]
      }
    ]
  };
  const mockProjectService = {
    listProjectsResponse: () => {
      const res: HttpResponse<Array<Project>> = new HttpResponse<Array<Project>>({
        headers: new HttpHeaders({'x-total-count': '3'}),
        body: [project1, project2, project3]
      });
      return of(res).pipe(delay(0));
    }
  };
  const fakedRobotService = {
    ListRobotResponse() {
      const res: HttpResponse<Array<Robot>> = new HttpResponse<Array<Robot>>({
        headers: new HttpHeaders({'x-total-count': '3'}),
        body: [robot1, robot2, robot3]
      });
      return of(res).pipe(delay(0));
    }
  };
  const fakedMessageHandlerService = {
    showSuccess() {
    },
    error() {
    }
  };
  beforeEach(waitForAsync(() => {
    TestBed.configureTestingModule({
      imports: [
        SharedTestingModule
      ],
      declarations: [ SystemRobotAccountsComponent ],
      providers: [
        { provide: MessageHandlerService, useValue: fakedMessageHandlerService },
        ConfirmationDialogService,
        OperationService,
        { provide: RobotService, useValue: fakedRobotService},
        { provide: ProjectService, useValue: mockProjectService},
      ],
      schemas: [
        NO_ERRORS_SCHEMA
      ],
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(SystemRobotAccountsComponent);
    component = fixture.componentInstance;
    component.searchSub = new Subscription();
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
  it('should render robot list',  async () => {
    fixture.autoDetectChanges();
    await fixture.whenStable();
    const rows = fixture.nativeElement.querySelectorAll('clr-dg-row');
    expect(rows.length).toEqual(3);
  });
});
