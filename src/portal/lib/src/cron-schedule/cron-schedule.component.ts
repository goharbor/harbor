import {
  Component,
  EventEmitter,
  Output,
  Input,
  OnChanges,
  SimpleChanges,
  SimpleChange
} from "@angular/core";
import { OriginCron } from "../service/interface";
import { cronRegex } from "../utils";
const SCHEDULE_TYPE = {
  NONE: "None",
  DAILY: "Daily",
  WEEKLY: "Weekly",
  HOURLY: "Hourly",
  CUSTOM: "Custom"
};
@Component({
  selector: "cron-selection",
  templateUrl: "./cron-schedule.component.html",
  styleUrls: ["./cron-schedule.component.scss"]
})
export class CronScheduleComponent implements OnChanges {
  @Input() originCron: OriginCron;
  @Input() labelEdit: string;
  @Input() labelCurrent: string;
  dateInvalid: boolean;
  originScheduleType: string;
  oriCron: string;
  cronString: string;
  isEditMode: boolean = false;
  SCHEDULE_TYPE = SCHEDULE_TYPE;
  scheduleType: string;
  @Output() inputvalue = new EventEmitter<string>();

  ngOnChanges(changes: SimpleChanges): void {
    let cronChange: SimpleChange = changes["originCron"];
    if (cronChange.currentValue) {
      this.originScheduleType = cronChange.currentValue.type;
      this.oriCron = cronChange.currentValue.cron;
    }
  }
  editSchedule() {
    if (!this.originScheduleType) {
      return;
    }
    this.isEditMode = true;
    this.scheduleType = this.originScheduleType;
    if (this.scheduleType && this.scheduleType === SCHEDULE_TYPE.CUSTOM) {
      this.cronString = this.oriCron;
      this.dateInvalid = !cronRegex(this.cronString);
    } else {
      this.cronString = "";
      this.dateInvalid = false;
    }
  }

  inputInvalid() {
    this.dateInvalid = !cronRegex(this.cronString);
  }

  blurInvalid() {
    if (!this.cronString) {
      this.dateInvalid = true;
    }
  }

  public resetSchedule() {
    this.originScheduleType = this.scheduleType;
    this.oriCron = this.cronString.replace(/\s+/g, " ").trim();
    this.isEditMode = false;
  }

  save(): void {
    if (this.scheduleType === SCHEDULE_TYPE.CUSTOM ) {
      if (this.cronString === '') {
        this.dateInvalid = true;
      }
      if (this.dateInvalid) {
        return;
      }
    }

    let scheduleTerm: string = "";
    if (this.scheduleType && this.scheduleType === SCHEDULE_TYPE.NONE) {
      scheduleTerm = "";
    } else if (this.scheduleType && this.scheduleType === SCHEDULE_TYPE.HOURLY) {
      scheduleTerm = "0 0 * * * *";
    } else if (this.scheduleType && this.scheduleType === SCHEDULE_TYPE.DAILY) {
      scheduleTerm = "0 0 0 * * *";
    } else if (this.scheduleType && this.scheduleType === SCHEDULE_TYPE.WEEKLY) {
      scheduleTerm = "0 0 0 * * 0";
    } else {
      scheduleTerm = this.cronString;
    }
    scheduleTerm = scheduleTerm.replace(/\s+/g, " ").trim();
    this.inputvalue.emit(scheduleTerm);
  }
}
