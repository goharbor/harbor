import { TestBed, inject } from '@angular/core/testing';

import { HelmChartService } from './helm-chart.service';

describe('HelmChartService', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [HelmChartService]
    });
  });

  it('should be created', inject([HelmChartService], (service: HelmChartService) => {
    expect(service).toBeTruthy();
  }));
});
