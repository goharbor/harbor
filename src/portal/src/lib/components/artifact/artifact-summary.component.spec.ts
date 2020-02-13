import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { ArtifactSummaryComponent } from "./artifact-summary.component";
import { of } from "rxjs";
import { Artifact } from "../../../../ng-swagger-gen/models/artifact";
import { ProjectService } from "../../services";
import { ArtifactService } from "../../../../ng-swagger-gen/services/artifact.service";
import { ErrorHandler } from "../../utils/error-handler";
import { ClarityModule } from "@clr/angular";
import { NO_ERRORS_SCHEMA } from "@angular/core";

describe('ArtifactSummaryComponent', () => {

  const mockedArtifact: Artifact = {
    id: 123,
    type: 'IMAGE'
  };

  const fakedProjectService = {
    getProject() {
      return of({name: 'test'});
    }
  };

  const fakedArtifactService = {
    getArtifact() {
       return of(mockedArtifact);
    }
  };
  let component: ArtifactSummaryComponent;
  let fixture: ComponentFixture<ArtifactSummaryComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [
        ClarityModule
      ],
      declarations: [
        ArtifactSummaryComponent
      ],
      schemas: [
        NO_ERRORS_SCHEMA
      ],
      providers: [
        {provide: ProjectService, useValue: fakedProjectService},
        {provide: ArtifactService, useValue: fakedArtifactService},
        ErrorHandler
      ]
    })
      .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ArtifactSummaryComponent);
    component = fixture.componentInstance;
    component.repositoryName = 'demo';
    component.artifactDigest = 'sha: acf4234f';
    fixture.detectChanges();
  });

  it('should create and get artifactDetails', async () => {
    expect(component).toBeTruthy();
    await fixture.whenStable();
    expect(component.artifact.type).toEqual('IMAGE');
  });
});
