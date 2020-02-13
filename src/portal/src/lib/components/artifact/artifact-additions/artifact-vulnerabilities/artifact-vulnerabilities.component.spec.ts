import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { ArtifactVulnerabilitiesComponent } from './artifact-vulnerabilities.component';

describe('ArtifactVulnerabilitiesComponent', () => {
  let component: ArtifactVulnerabilitiesComponent;
  let fixture: ComponentFixture<ArtifactVulnerabilitiesComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ ArtifactVulnerabilitiesComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ArtifactVulnerabilitiesComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
