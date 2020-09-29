import { TestBed } from '@angular/core/testing';
import { HttpClient } from '@angular/common/http';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { OidcOnboardService } from './oidc-onboard.service';

describe('OidcOnboardService', () => {
  beforeEach(() => TestBed.configureTestingModule({
    imports: [
      HttpClientTestingModule
    ],
    providers: [
      OidcOnboardService
    ]
  }));

  it('should be created', () => {
    const service: OidcOnboardService = TestBed.get(OidcOnboardService);
    expect(service).toBeTruthy();
  });
});
