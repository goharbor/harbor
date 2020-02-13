import { TestBed, inject } from '@angular/core/testing';
import { RouterTestingModule } from '@angular/router/testing';
import { LeavingArtifactSummaryRouteDeactivate } from './leaving-artifact-summary-deactivate.service';
import { ConfirmationDialogService } from '../confirmation-dialog/confirmation-dialog.service';
import { of } from 'rxjs';

describe('LeavingArtifactSummaryRouteDeactivate', () => {
  let fakeConfirmationDialogService = {
    confirmationConfirm$: of({
      state: 1,
      source: 2
    })
  };
  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [
        RouterTestingModule
      ],
      providers: [
        LeavingArtifactSummaryRouteDeactivate,
        { provide: ConfirmationDialogService, useValue: fakeConfirmationDialogService }
      ]
    });
  });

  it('should be created', inject([LeavingArtifactSummaryRouteDeactivate], (service: LeavingArtifactSummaryRouteDeactivate) => {
    expect(service).toBeTruthy();
  }));
});
