import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { ArtifactTagComponent } from './artifact-tag.component';

describe('ArtifactTagComponent', () => {
  let component: ArtifactTagComponent;
  let fixture: ComponentFixture<ArtifactTagComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ ArtifactTagComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ArtifactTagComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
