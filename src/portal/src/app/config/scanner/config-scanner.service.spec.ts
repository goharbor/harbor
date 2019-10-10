import { TestBed, inject } from '@angular/core/testing';
import { SharedModule } from "../../shared/shared.module";
import { ConfigScannerService } from "./config-scanner.service";

describe('TagService', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [
        SharedModule
      ],
      providers: [
        ConfigScannerService
      ]
    });
  });

  it('should be initialized', inject([ConfigScannerService], (service: ConfigScannerService) => {
    expect(service).toBeTruthy();
  }));
});
