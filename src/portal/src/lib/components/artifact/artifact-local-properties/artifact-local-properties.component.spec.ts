import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { ArtifactLocalPropertiesComponent } from './artifact-local-properties.component';

describe('ArtifactLocalPropertiesComponent', () => {
  let component: ArtifactLocalPropertiesComponent;
  let fixture: ComponentFixture<ArtifactLocalPropertiesComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ ArtifactLocalPropertiesComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ArtifactLocalPropertiesComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
