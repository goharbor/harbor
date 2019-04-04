import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { GcPageComponent } from './gc-page.component';

describe('GcPageComponent', () => {
  let component: GcPageComponent;
  let fixture: ComponentFixture<GcPageComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ GcPageComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(GcPageComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
