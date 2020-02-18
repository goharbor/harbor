import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { ErrorHandler } from "../../../utils/error-handler/error-handler";

import { ArtifactTagComponent } from './artifact-tag.component';
import { SharedModule } from '../../../utils/shared/shared.module';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { OperationService } from '../../operation/operation.service';
import { TagService } from '../../../services';
import { of } from 'rxjs';
import { SERVICE_CONFIG, IServiceConfig } from '../../../entities/service.config';

describe('ArtifactTagComponent', () => {
  let component: ArtifactTagComponent;
  let fixture: ComponentFixture<ArtifactTagComponent>;
  const mockErrorHandler = {
    error: () => {}
  };
  const mockTagService = {
    newTag: () => of([]),
    deleteTag: () => of(null),
  };
  const config: IServiceConfig = {
    repositoryBaseEndpoint: "/api/repositories/testing"
  };
  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [
        SharedModule,
        BrowserAnimationsModule,
        HttpClientTestingModule
      ],
      schemas: [
        CUSTOM_ELEMENTS_SCHEMA
      ],
      declarations: [ ArtifactTagComponent ],
      providers: [
        ErrorHandler,
        { provide: SERVICE_CONFIG, useValue: config },
        { provide: mockErrorHandler, useValue: ErrorHandler },
        { provide: TagService, useValue: mockTagService },
        { provide: OperationService },
      ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ArtifactTagComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
