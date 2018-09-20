import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { GcComponent } from './gc.component';

describe('GcComponent', () => {
  let component: GcComponent;
  let fixture: ComponentFixture<GcComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ GcComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(GcComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
