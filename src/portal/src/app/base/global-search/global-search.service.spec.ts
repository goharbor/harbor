import { TestBed, inject, getTestBed } from '@angular/core/testing';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { GlobalSearchService } from './global-search.service';
import { Injector } from '@angular/core';
import { SearchResults } from './search-results';

describe('GlobalSearchService', () => {
  let injector: TestBed;
  let service: GlobalSearchService;
  let httpMock: HttpTestingController;


  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [GlobalSearchService],
      imports: [
        HttpClientTestingModule
      ]
    });
    injector = getTestBed();
    service = injector.get(GlobalSearchService);
    httpMock = injector.get(HttpTestingController);

  });

  it('should be created', inject([GlobalSearchService], (service1: GlobalSearchService) => {
    expect(service1).toBeTruthy();
  }));
  it('doSearch should return data', () => {
    service.doSearch("library").subscribe((res) => {
      expect(res).toEqual(new SearchResults());
    });

    const req = httpMock.expectOne('/api/search?q=library');
    expect(req.request.method).toBe('GET');
    req.flush(new SearchResults());
  });
});
