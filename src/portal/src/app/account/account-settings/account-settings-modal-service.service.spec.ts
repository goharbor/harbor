import { TestBed } from '@angular/core/testing';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { AccountSettingsModalService } from './account-settings-modal-service.service';

describe('AccountSettingsModalServiceService', () => {
  beforeEach(() => TestBed.configureTestingModule({
    imports: [
      HttpClientTestingModule
    ],
    providers: [
      AccountSettingsModalService
    ]
  }));

  it('should be created', () => {
    const service: AccountSettingsModalService = TestBed.get(AccountSettingsModalService);
    expect(service).toBeTruthy();
  });
});
