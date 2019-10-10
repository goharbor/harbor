import { TestBed, inject } from '@angular/core/testing';

import { ConfirmationDialogService } from './confirmation-dialog.service';

describe('ConfirmationDialogService', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [ConfirmationDialogService]
    });
  });

  it('should be created', inject([ConfirmationDialogService], (service: ConfirmationDialogService) => {
    expect(service).toBeTruthy();
  }));
});
