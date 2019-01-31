import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { AddRobotComponent } from './add-robot.component';

describe('AddRobotComponent', () => {
  let component: AddRobotComponent;
  let fixture: ComponentFixture<AddRobotComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ AddRobotComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(AddRobotComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
