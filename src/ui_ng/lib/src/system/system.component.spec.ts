import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { HttpModule } from '@angular/http';

import { SystemComponent } from './system.component';
import { SystemInfoService } from './providers/system-info.service';
import { SERVICE_CONFIG, IServiceConfig } from '../service.config';

export const testConfig: IServiceConfig = {
  systemInfoEndpoint: "/api/systeminfo",
  repositoryBaseEndpoint: "",
  logBaseEndpoint: "",
  targetBaseEndpoint: "",
  replicationRuleEndpoint: "",
  replicationJobEndpoint: ""
};

describe('SystemComponent', () => {
  let component: SystemComponent;
  let fixture: ComponentFixture<SystemComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [HttpModule],
      declarations: [SystemComponent],
      providers: [
        { provide: SERVICE_CONFIG, useValue: testConfig },
        SystemInfoService
      ]
    })
      .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(SystemComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should be created', () => {
    expect(component).toBeTruthy();
  });
});
