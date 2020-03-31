import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { ArtifactTagComponent } from './artifact-tag.component';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { of } from 'rxjs';
import { IServiceConfig, SERVICE_CONFIG } from "../../../../../lib/entities/service.config";
import { SharedModule } from "../../../../../lib/utils/shared/shared.module";
import { ErrorHandler } from "../../../../../lib/utils/error-handler";
import { ArtifactService } from '../../../../../../ng-swagger-gen/services/artifact.service';
import { OperationService } from "../../../../../lib/components/operation/operation.service";
import { CURRENT_BASE_HREF } from "../../../../../lib/utils/utils";
import { USERSTATICPERMISSION, UserPermissionService, UserPermissionDefaultService } from '../../../../../lib/services';
import { TagService } from '../../../../../../ng-swagger-gen/services/tag.service';
import { delay } from 'rxjs/operators';


describe('ArtifactTagComponent', () => {
  let component: ArtifactTagComponent;
  let fixture: ComponentFixture<ArtifactTagComponent>;
  const mockErrorHandler = {
    error: () => {}
  };
  const mockTagService = {
    listTagsResponse: () => of({headers: null, body: []}).pipe(delay(0)),
    listTags: () => of([]),
  };
  const mockArtifactService = {
    createTag: () => of([]),
    deleteTag: () => of(null),
  };
  const config: IServiceConfig = {
    repositoryBaseEndpoint: CURRENT_BASE_HREF + "/repositories/testing"
  };
  let userPermissionService;
  const permissions = [
    { resource: USERSTATICPERMISSION.REPOSITORY_TAG.KEY, action: USERSTATICPERMISSION.REPOSITORY_TAG.VALUE.DELETE },
    { resource: USERSTATICPERMISSION.REPOSITORY_TAG.KEY, action: USERSTATICPERMISSION.REPOSITORY_TAG.VALUE.CREATE },
  ];
  let mockHasDeleteImagePermission: boolean = true;
  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [
        SharedModule,
        BrowserAnimationsModule,
        HttpClientTestingModule
      ],
      schemas: [
        CUSTOM_ELEMENTS_SCHEMA
      ],
      declarations: [ ArtifactTagComponent ],
      providers: [
        ErrorHandler,
        { provide: SERVICE_CONFIG, useValue: config },
        { provide: mockErrorHandler, useValue: ErrorHandler },
        { provide: ArtifactService, useValue: mockArtifactService },
        { provide: TagService, useValue: mockTagService },
        { provide: UserPermissionService, useClass: UserPermissionDefaultService },
        { provide: OperationService },
      ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ArtifactTagComponent);
    component = fixture.componentInstance;
    userPermissionService = fixture.debugElement.injector.get(UserPermissionService);
    spyOn(userPermissionService, "hasProjectPermissions")
      .withArgs(component.projectId, permissions)
      .and.returnValue(of([
        mockHasDeleteImagePermission]));
    component.artifactDetails = {id: 1};
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
