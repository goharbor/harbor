import { TestBed, inject } from '@angular/core/testing';

import { TopRepoService } from './top-repository.service';

xdescribe('TopRepoService', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [TopRepoService]
    });
  });

  it('should be created', inject([TopRepoService], (service: TopRepoService) => {
    expect(service).toBeTruthy();
  }));
});
