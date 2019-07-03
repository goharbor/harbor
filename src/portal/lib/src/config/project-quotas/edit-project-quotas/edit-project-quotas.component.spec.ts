import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { EditProjectQuotasComponent } from './edit-project-quotas.component';

describe('EditProjectQuotasComponent', () => {
  let component: EditProjectQuotasComponent;
  let fixture: ComponentFixture<EditProjectQuotasComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ EditProjectQuotasComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(EditProjectQuotasComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
