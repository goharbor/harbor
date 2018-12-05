import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { DevCenterComponent } from './dev-center.component';

describe('DevCenterComponent', () => {
  let component: DevCenterComponent;
  let fixture: ComponentFixture<DevCenterComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ DevCenterComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(DevCenterComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

});
