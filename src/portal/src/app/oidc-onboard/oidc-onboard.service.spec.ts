import { TestBed } from '@angular/core/testing';

import { OidcOnboardService } from './oidc-onboard.service';

describe('OidcOnboardService', () => {
  beforeEach(() => TestBed.configureTestingModule({}));

  it('should be created', () => {
    const service: OidcOnboardService = TestBed.get(OidcOnboardService);
    expect(service).toBeTruthy();
  });
});
