import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { ProjectQuotasComponent } from './project-quotas.component';

describe('ProjectQuotasComponent', () => {
  let component: ProjectQuotasComponent;
  let fixture: ComponentFixture<ProjectQuotasComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ ProjectQuotasComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ProjectQuotasComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
