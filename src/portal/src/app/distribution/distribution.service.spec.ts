import { TestBed, inject } from '@angular/core/testing';
import { HttpClientTestingModule } from '@angular/common/http/testing';

import { DistributionService } from './distribution.service';

describe('DistributionService', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
      providers: [DistributionService]
    });
  });

  it('should be created', inject(
    [DistributionService],
    (service: DistributionService) => {
      expect(service).toBeTruthy();
    }
  ));
});
