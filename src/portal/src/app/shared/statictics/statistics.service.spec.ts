import { TestBed, inject } from '@angular/core/testing';

import { StatisticsService } from './statistics.service';

xdescribe('StatisticsService', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [StatisticsService]
    });
  });

  it('should be created', inject([StatisticsService], (service: StatisticsService) => {
    expect(service).toBeTruthy();
  }));
});
