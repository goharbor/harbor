import { TestBed, inject } from '@angular/core/testing';

import { ConfigurationService } from './config.service';

xdescribe('ConfigService', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [ConfigurationService]
    });
  });

  it('should be created', inject([ConfigurationService], (service: ConfigurationService) => {
    expect(service).toBeTruthy();
  }));
});
