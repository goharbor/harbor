import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { ArtifactCommonPropertiesComponent } from './artifact-common-properties.component';

describe('ArtifactCommonPropertiesComponent', () => {
  let component: ArtifactCommonPropertiesComponent;
  let fixture: ComponentFixture<ArtifactCommonPropertiesComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ ArtifactCommonPropertiesComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ArtifactCommonPropertiesComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
