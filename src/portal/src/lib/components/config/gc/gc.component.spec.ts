import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { GcComponent } from './gc.component';
import { SERVICE_CONFIG, IServiceConfig } from '../../../entities/service.config';
import { GcApiRepository, GcApiDefaultRepository} from './gc.api.repository';
import { GcRepoService } from './gc.service';
import { SharedModule } from "../../../utils/shared/shared.module";
import { ErrorHandler } from '../../../utils/error-handler/error-handler';
import { GcViewModelFactory } from './gc.viewmodel.factory';
import { CronScheduleComponent } from '../../cron-schedule/cron-schedule.component';
import { CronTooltipComponent } from "../../cron-schedule/cron-tooltip/cron-tooltip.component";
import { of } from 'rxjs';
import { GcJobData } from './gcLog';

describe('GcComponent', () => {
  let component: GcComponent;
  let fixture: ComponentFixture<GcComponent>;
  let gcRepoService: GcRepoService;
  let config: IServiceConfig = {
    systemInfoEndpoint: "/api/system/gc"
  };
  let mockSchedule = [];
  let mockJobs: GcJobData[] = [
    {
    id: 22222,
    schedule: null,
    job_status: 'string',
    creation_time: new Date().toDateString(),
    update_time: new Date().toDateString(),
    job_name: 'string',
    job_kind: 'string',
    job_uuid: 'string',
    delete: false
    }
  ];
  let spySchedule: jasmine.Spy;
  let spyJobs: jasmine.Spy;
  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [
        SharedModule
      ],
      declarations: [ GcComponent,  CronScheduleComponent, CronTooltipComponent],
      providers: [
        { provide: GcApiRepository, useClass: GcApiDefaultRepository },
        { provide: SERVICE_CONFIG, useValue: config },
        GcRepoService,
        ErrorHandler,
        GcViewModelFactory
      ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(GcComponent);
    component = fixture.componentInstance;

    gcRepoService = fixture.debugElement.injector.get(GcRepoService);
    spySchedule = spyOn(gcRepoService, "getSchedule").and.returnValues(of(mockSchedule));
    spyJobs = spyOn(gcRepoService, "getJobs").and.returnValues(of(mockJobs));
    fixture.detectChanges();
  });
  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
