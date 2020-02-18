import { TestBed } from '@angular/core/testing';
import { BaseHrefInterceptService } from "./base-href-intercept.service";


describe('BaseHrefSwitchService', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [
        BaseHrefInterceptService
      ]
    });
  });

  it('should be created', () => {
    const service: BaseHrefInterceptService = TestBed.get(BaseHrefInterceptService);
    expect(service).toBeTruthy();
  });
});
