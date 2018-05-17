import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { AddGroupModalComponent } from './add-group-modal.component';

describe('AddGroupModalComponent', () => {
  let component: AddGroupModalComponent;
  let fixture: ComponentFixture<AddGroupModalComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ AddGroupModalComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(AddGroupModalComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
