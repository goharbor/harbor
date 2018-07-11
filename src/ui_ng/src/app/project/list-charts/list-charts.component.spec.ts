import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { ListChartsComponent } from './list-charts.component';

describe('ListChartsComponent', () => {
  let component: ListChartsComponent;
  let fixture: ComponentFixture<ListChartsComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ ListChartsComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ListChartsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
