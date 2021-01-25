import { TestBed, inject } from '@angular/core/testing';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { StatisticsService } from './statistics.service';

describe('StatisticsService', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [
        HttpClientTestingModule
      ],
      providers: [StatisticsService]
    });
  });

  it('should be created', inject([StatisticsService], (service: StatisticsService) => {
    expect(service).toBeTruthy();
  }));
});
