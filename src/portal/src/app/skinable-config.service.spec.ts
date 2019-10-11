import { TestBed, inject } from '@angular/core/testing';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { SkinableConfig } from './skinable-config.service';

describe('SkinableConfig', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [
        HttpClientTestingModule
      ],
      providers: [SkinableConfig]
    });
  });

  it('should be created', inject([SkinableConfig], (service: SkinableConfig) => {
    expect(service).toBeTruthy();
  }));
});
