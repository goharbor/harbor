import { waitForAsync, ComponentFixture, TestBed } from '@angular/core/testing';
import { GcComponent } from './gc.component';
import { GcApiRepository, GcApiDefaultRepository} from './gc.api.repository';
import { GcRepoService } from './gc.service';
import { ErrorHandler } from '../../../../shared/units/error-handler';
import { GcViewModelFactory } from './gc.viewmodel.factory';
import { CronScheduleComponent } from '../../../../shared/components/cron-schedule';
import { CronTooltipComponent } from "../../../../shared/components/cron-schedule";
import { of } from 'rxjs';
import { GcJobData } from './gcLog';
import { CURRENT_BASE_HREF } from "../../../../shared/units/utils";
import { SharedTestingModule } from "../../../../shared/shared.module";

describe('GcComponent', () => {
  let component: GcComponent;
  let fixture: ComponentFixture<GcComponent>;
  let gcRepoService: GcRepoService;
  let mockSchedule = [];
  let mockJobs: GcJobData[] = [
    {
    id: 22222,
    schedule: null,
    job_status: 'string',
    job_parameters: '{"dry_run":true}',
    creation_time: new Date().toDateString(),
    update_time: new Date().toDateString(),
    job_name: 'string',
    job_kind: 'string',
    job_uuid: 'string',
    delete: false
    }
  ];
  const fakedErrorHandler = {
    error(error) {
      return error;
    },
    info(info) {
      return info;
    }
  };
  let spySchedule: jasmine.Spy;
  let spyJobs: jasmine.Spy;
  let spyGcNow: jasmine.Spy;
  beforeEach(waitForAsync(() => {
    TestBed.configureTestingModule({
      imports: [
        SharedTestingModule
      ],
      declarations: [ GcComponent,  CronScheduleComponent, CronTooltipComponent],
      providers: [
        { provide: GcApiRepository, useClass: GcApiDefaultRepository },
        { provide: ErrorHandler, useValue: fakedErrorHandler },
        GcRepoService,
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
    spyGcNow = spyOn(gcRepoService, "manualGc").and.returnValues(of(true));
    fixture.detectChanges();
  });
  it('should create', () => {
    expect(component).toBeTruthy();
  });
  it('should get schedule and job', () => {
    expect(spySchedule.calls.count()).toEqual(1);
    expect(spyJobs.calls.count()).toEqual(1);
  });
  it('should trigger gcNow', () => {
    const ele: HTMLButtonElement = fixture.nativeElement.querySelector('.gc-start-btn');
    ele.click();
    fixture.detectChanges();
    expect(spyGcNow.calls.count()).toEqual(1);
  });
});
