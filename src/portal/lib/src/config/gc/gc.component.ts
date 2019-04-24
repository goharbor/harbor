import { Component, Input, Output, EventEmitter, ViewChild, OnInit } from '@angular/core';
import { TranslateService } from '@ngx-translate/core';
import { GcJobViewModel, WeekDay } from "./gcLog";
import { GcViewModelFactory } from "./gc.viewmodel.factory";
import { GcRepoService } from "./gc.service";
import { WEEKDAYS, SCHEDULE_TYPE, ONE_MINITUE, THREE_SECONDS} from './gc.const';
import { GcUtility } from './gc.utility';
import { ErrorHandler } from '../../error-handler/index';

@Component({
  selector: 'gc-config',
  templateUrl: './gc.component.html',
  styleUrls: ['./gc.component.scss']
})
export class GcComponent implements OnInit {
  jobs: Array<GcJobViewModel> = [];
  schedule: any;
  originScheduleType: string;
  originOffTime: any = { value: null, text: "" };
  originWeekDay: any = { value: null, text: "" };
  scheduleType: string;
  isEditMode: boolean = false;
  weekDays = WEEKDAYS;
  SCHEDULE_TYPE = SCHEDULE_TYPE;
  weekDay: WeekDay = WEEKDAYS[0];
  dailyTime: string;
  disableGC: boolean = false;

  constructor(private gcRepoService: GcRepoService,
    private gcViewModelFactory: GcViewModelFactory,
    private gcUtility: GcUtility,
    private errorHandler: ErrorHandler,
    private translate: TranslateService) {
    translate.setDefaultLang('en-us');
  }


  ngOnInit() {
    this.getCurrentSchedule();
    this.getJobs();
  }

  getCurrentSchedule() {
    this.gcRepoService.getSchedule().subscribe(schedule => {
      this.initSchedule(schedule);
    });
  }

  private initSchedule(schedule: any) {
    if (schedule && schedule.length > 0) {
      this.schedule = schedule[0];
      const cron = this.schedule.schedule;
      this.originScheduleType = cron.type;
      this.originWeekDay = this.weekDays[cron.weekday];
      let dailyTime = this.gcUtility.getDailyTime(cron.offtime);
      this.originOffTime = { value: cron.offtime, text: dailyTime };
    } else {
      this.originScheduleType = SCHEDULE_TYPE.NONE;
    }
  }

  editSchedule() {
    this.isEditMode = true;
    this.scheduleType = this.originScheduleType;
    if (this.originWeekDay.value) {
      this.weekDay = this.originWeekDay;
    } else {
      this.weekDay = this.weekDays[0];
    }

    if (this.originOffTime.value) {
      this.dailyTime = this.originOffTime.text;
    } else {
      this.dailyTime = "00:00";
    }
  }

  getJobs() {
    this.gcRepoService.getJobs().subscribe(jobs => {
      this.jobs = this.gcViewModelFactory.createJobViewModel(jobs);
    });
  }

  gcNow(): void {
    this.disableGC = true;
    setTimeout(() => {this.enableGc(); }, ONE_MINITUE);

    this.gcRepoService.manualGc().subscribe(response => {
      this.translate.get('GC.MSG_SUCCESS').subscribe((res: string) => {
        this.errorHandler.info(res);
      });
      this.getJobs();
      setTimeout(() => {this.getJobs(); }, THREE_SECONDS); // to avoid some jobs not finished.
    }, error => {
      this.errorHandler.error(error);
    });
  }

  private enableGc () {
    this.disableGC = false;
  }

  private resetSchedule(offTime) {
    this.schedule = {
      schedule: {
        type: this.scheduleType,
        offTime: offTime,
        weekDay: this.weekDay.value
      }
    };
    this.originScheduleType = this.scheduleType;
    this.originWeekDay = this.weekDay;
    this.originOffTime = { value: offTime, text: this.dailyTime };
    this.isEditMode = false;
    this.getJobs();
  }

  logLink(id) {
    return this.gcRepoService.getLogLink(id);
  }

  scheduleGc(): void {
    let offTime = this.gcUtility.getOffTime(this.dailyTime);
    let schedule = this.schedule;
    if (schedule && schedule.schedule && schedule.schedule.type !== SCHEDULE_TYPE.NONE) {
      this.gcRepoService.putScheduleGc(this.scheduleType, offTime, this.weekDay.value).subscribe(response => {
        this.translate.get('GC.MSG_SCHEDULE_RESET').subscribe((res: string) => {
          this.errorHandler.info(res);
        });
        this.resetSchedule(offTime);
      }, error => {
        this.errorHandler.error(error);
      });
    } else {
      this.gcRepoService.postScheduleGc(this.scheduleType, offTime, this.weekDay.value).subscribe(response => {
        this.translate.get('GC.MSG_SCHEDULE_SET').subscribe((res: string) => {
          this.errorHandler.info(res);
        });
        this.resetSchedule(offTime);
      }, error => {
        this.errorHandler.error(error);
      });
    }
  }
}
