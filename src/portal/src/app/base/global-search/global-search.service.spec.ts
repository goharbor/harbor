import { TestBed, inject } from '@angular/core/testing';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { GlobalSearchService } from './global-search.service';

describe('GlobalSearchService', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [GlobalSearchService],
      imports: [
        HttpClientTestingModule
      ]
    });
  });

  it('should be created', inject([GlobalSearchService], (service: GlobalSearchService) => {
    expect(service).toBeTruthy();
  }));
});
