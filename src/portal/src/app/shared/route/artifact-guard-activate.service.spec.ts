import { TestBed } from '@angular/core/testing';

import { ArtifactGuardActivateService } from './artifact-guard-activate.service';

describe('ArtifactGuardActivateService', () => {
  beforeEach(() => TestBed.configureTestingModule({}));

  it('should be created', () => {
    const service: ArtifactGuardActivateService = TestBed.get(ArtifactGuardActivateService);
    expect(service).toBeTruthy();
  });
});
