import { TestBed } from '@angular/core/testing';

import { AccountSettingsModalService } from './account-settings-modal-service.service';

describe('AccountSettingsModalServiceService', () => {
  beforeEach(() => TestBed.configureTestingModule({}));

  it('should be created', () => {
    const service: AccountSettingsModalService = TestBed.get(AccountSettingsModalService);
    expect(service).toBeTruthy();
  });
});
