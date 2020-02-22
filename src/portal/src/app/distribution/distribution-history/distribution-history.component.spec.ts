import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { TranslateModule } from '@ngx-translate/core';
import { ClarityModule } from '@clr/angular';
import { SharedModule } from '../../shared/shared.module';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { HttpClientTestingModule } from '@angular/common/http/testing';

import { DistributionHistoryComponent } from './distribution-history.component';
import { DistributionService } from '../distribution.service';

describe('DistributionHistoryComponent', () => {
  let component: DistributionHistoryComponent;
  let fixture: ComponentFixture<DistributionHistoryComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [
        ClarityModule,
        TranslateModule,
        SharedModule,
        HttpClientTestingModule
      ],
      schemas: [CUSTOM_ELEMENTS_SCHEMA],
      providers: [DistributionService],
      declarations: [DistributionHistoryComponent]
    }).compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(DistributionHistoryComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  // create
  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
