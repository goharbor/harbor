import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { HelmChartDetailComponent } from './chart-detail.component';

describe('ChartDetailComponent', () => {
  let component: HelmChartDetailComponent;
  let fixture: ComponentFixture<HelmChartDetailComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ HelmChartDetailComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(HelmChartDetailComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
