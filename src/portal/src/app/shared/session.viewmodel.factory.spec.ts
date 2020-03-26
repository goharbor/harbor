import { TestBed } from '@angular/core/testing';

import { SessionViewmodelFactory } from './session.viewmodel.factory';

describe('SessionViewmodelFactory', () => {
  beforeEach(() => TestBed.configureTestingModule({}));

  it('should be created', () => {
    const service: SessionViewmodelFactory = TestBed.get(SessionViewmodelFactory);
    expect(service).toBeTruthy();
  });
});
