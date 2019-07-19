import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { AddHttpAuthGroupComponent } from './add-http-auth-group.component';

describe('AddHttpAuthGroupComponent', () => {
  let component: AddHttpAuthGroupComponent;
  let fixture: ComponentFixture<AddHttpAuthGroupComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ AddHttpAuthGroupComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(AddHttpAuthGroupComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
