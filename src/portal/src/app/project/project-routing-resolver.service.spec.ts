import { TestBed, inject } from '@angular/core/testing';

import { ProjectRoutingResolver } from './project-routing-resolver.service';

xdescribe('ProjectRoutingResolverService', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [ProjectRoutingResolver]
    });
  });

  it('should be created', inject([ProjectRoutingResolver], (service: ProjectRoutingResolver) => {
    expect(service).toBeTruthy();
  }));
});
