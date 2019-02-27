import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { RobotAccountComponent } from './robot-account.component';

describe('RobotAccountComponent', () => {
  let component: RobotAccountComponent;
  let fixture: ComponentFixture<RobotAccountComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ RobotAccountComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(RobotAccountComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
