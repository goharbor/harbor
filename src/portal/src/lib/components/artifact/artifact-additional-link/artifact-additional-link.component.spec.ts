import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { ArtifactAdditionalLinkComponent } from './artifact-additional-link.component';

describe('ArtifactAdditionalLinkComponent', () => {
  let component: ArtifactAdditionalLinkComponent;
  let fixture: ComponentFixture<ArtifactAdditionalLinkComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ ArtifactAdditionalLinkComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ArtifactAdditionalLinkComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
