import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { ClarityModule } from '@clr/angular';
import { SharedModule } from '../../shared/shared.module';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { DistributionInstancesComponent } from './distribution-instances.component';
import { MsgChannelService } from '../msg-channel.service';
import { PreheatService } from "../../../../ng-swagger-gen/services/preheat.service";

describe('DistributionInstanceComponent', () => {
  let component: DistributionInstancesComponent;
  let fixture: ComponentFixture<DistributionInstancesComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [
        ClarityModule,
        TranslateModule,
        SharedModule,
        HttpClientTestingModule
      ],
      schemas: [CUSTOM_ELEMENTS_SCHEMA],
      providers: [PreheatService, MsgChannelService],
      declarations: [DistributionInstancesComponent]
    }).compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(DistributionInstancesComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
