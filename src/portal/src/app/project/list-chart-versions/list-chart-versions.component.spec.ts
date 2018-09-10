import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { ListChartVersionsComponent } from './list-chart-versions.component';

describe('ListChartVersionsComponent', () => {
  let component: ListChartVersionsComponent;
  let fixture: ComponentFixture<ListChartVersionsComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ ListChartVersionsComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ListChartVersionsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
