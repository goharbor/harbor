import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { ArtifactVulnerabilitiesComponent } from './artifact-vulnerabilities.component';
import { SharedModule } from '../../../../utils/shared/shared.module';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { ClarityModule } from '@clr/angular';
import { ErrorHandler } from '../../../../utils/error-handler';
import { AdditionsService } from '../additions.service';
import { of } from 'rxjs';
import { SERVICE_CONFIG, IServiceConfig } from '../../../../entities/service.config';

describe('ArtifactVulnerabilitiesComponent', () => {
  let component: ArtifactVulnerabilitiesComponent;
  let fixture: ComponentFixture<ArtifactVulnerabilitiesComponent>;
  const mockErrorHandler = {
    error: () => { }
  };
  const mockAdditionsService = {
    getDetailByLink: () => of([])
  };
  const config: IServiceConfig = {
    repositoryBaseEndpoint: "/api/repositories/testing"
  };
  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [
        SharedModule,
        BrowserAnimationsModule,
        HttpClientTestingModule,
        ClarityModule
      ],
      schemas: [
        CUSTOM_ELEMENTS_SCHEMA
      ],
      declarations: [ArtifactVulnerabilitiesComponent],
      providers: [
        { provide: SERVICE_CONFIG, useValue: config },
        {
          provide: ErrorHandler, useValue: mockErrorHandler
        },
        { provide: AdditionsService, useValue: mockAdditionsService },
      ]
    })
      .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ArtifactVulnerabilitiesComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
