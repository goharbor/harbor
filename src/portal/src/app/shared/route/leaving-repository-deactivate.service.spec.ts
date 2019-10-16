import { TestBed, inject } from '@angular/core/testing';
import { RouterTestingModule } from '@angular/router/testing';
import { LeavingRepositoryRouteDeactivate } from './leaving-repository-deactivate.service';
import { ConfirmationDialogService } from '../confirmation-dialog/confirmation-dialog.service';
import { of } from 'rxjs';

describe('LeavingRepositoryRouteDeactivate', () => {
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
        LeavingRepositoryRouteDeactivate,
        { provide: ConfirmationDialogService, useValue: fakeConfirmationDialogService }
      ]
    });
  });

  it('should be created', inject([LeavingRepositoryRouteDeactivate], (service: LeavingRepositoryRouteDeactivate) => {
    expect(service).toBeTruthy();
  }));
});
