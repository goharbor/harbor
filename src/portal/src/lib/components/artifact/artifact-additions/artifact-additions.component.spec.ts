import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { ArtifactAdditionsComponent } from './artifact-additions.component';

describe('ArtifactAdditionsComponent', () => {
  let component: ArtifactAdditionsComponent;
  let fixture: ComponentFixture<ArtifactAdditionsComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ ArtifactAdditionsComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ArtifactAdditionsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
