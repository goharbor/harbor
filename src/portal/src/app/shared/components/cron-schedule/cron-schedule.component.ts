import {
    Component,
    EventEmitter,
    Output,
    Input,
    OnChanges,
    SimpleChanges,
    SimpleChange,
} from '@angular/core';
import { OriginCron } from '../../services/interface';
import { cronRegex } from '../../units/utils';
import { TranslateService } from '@ngx-translate/core';
import { ErrorHandler } from '../../units/error-handler/error-handler';
const SCHEDULE_TYPE = {
    NONE: 'None',
    DAILY: 'Daily',
    WEEKLY: 'Weekly',
    HOURLY: 'Hourly',
    CUSTOM: 'Custom',
};
const PREFIX: string = '0 ';
@Component({
    selector: 'cron-selection',
    templateUrl: './cron-schedule.component.html',
    styleUrls: ['./cron-schedule.component.scss'],
})
export class CronScheduleComponent implements OnChanges {
    @Input() isInlineModel: boolean = false;
    @Input() originCron: OriginCron;
    @Input() labelEdit: string;
    @Input() labelCurrent: string;
    @Input() disabled: boolean;
    @Input() labelWidth: string = '200px';
    dateInvalid: boolean;
    originScheduleType: string;
    oriCron: string;
    cronString: string;
    isEditMode: boolean = false;
    SCHEDULE_TYPE = SCHEDULE_TYPE;
    scheduleType: string;
    @Output() inputvalue = new EventEmitter<string>();
    constructor(
        private translate: TranslateService,
        private errorHandler: ErrorHandler
    ) {}

    ngOnChanges(changes: SimpleChanges): void {
        let cronChange: SimpleChange = changes['originCron'];
        if (cronChange.currentValue) {
            this.originScheduleType = cronChange.currentValue.type;
            this.oriCron = cronChange.currentValue.cron;
        }
    }
    editSchedule() {
        if (!this.originScheduleType) {
            this.translate
                .get('SCHEDULE.NOSCHEDULE')
                .subscribe(res => this.errorHandler.error(res));
            return;
        }
        this.isEditMode = true;
        this.scheduleType = this.originScheduleType;
        if (this.scheduleType && this.scheduleType === SCHEDULE_TYPE.CUSTOM) {
            this.cronString = this.oriCron || PREFIX;
            this.dateInvalid = !cronRegex(this.cronString);
        } else {
            this.cronString = PREFIX;
            this.dateInvalid = false;
        }
    }

    inputInvalid(e: any) {
        this.dateInvalid = !cronRegex(this.cronString);
        this.setPrefix(e);
    }

    blurInvalid() {
        if (!this.cronString) {
            this.dateInvalid = true;
        }
    }

    public resetSchedule() {
        this.originScheduleType = this.scheduleType;
        this.oriCron = this.cronString.replace(/\s+/g, ' ').trim();
        this.isEditMode = false;
    }

    save(): void {
        if (this.scheduleType === SCHEDULE_TYPE.CUSTOM) {
            if (this.cronString === '') {
                this.dateInvalid = true;
            }
            if (this.dateInvalid) {
                return;
            }
        }

        let scheduleTerm: string = '';
        if (this.scheduleType && this.scheduleType === SCHEDULE_TYPE.NONE) {
            scheduleTerm = '';
        } else if (
            this.scheduleType &&
            this.scheduleType === SCHEDULE_TYPE.HOURLY
        ) {
            scheduleTerm = '0 0 * * * *';
        } else if (
            this.scheduleType &&
            this.scheduleType === SCHEDULE_TYPE.DAILY
        ) {
            scheduleTerm = '0 0 0 * * *';
        } else if (
            this.scheduleType &&
            this.scheduleType === SCHEDULE_TYPE.WEEKLY
        ) {
            scheduleTerm = '0 0 0 * * 0';
        } else {
            scheduleTerm = this.cronString;
        }
        scheduleTerm = scheduleTerm.replace(/\s+/g, ' ').trim();
        this.inputvalue.emit(scheduleTerm);
    }
    // set prefix '0 ', so user can not set item of 'seconds'
    setPrefix(e: any) {
        if (e && e.target) {
            if (
                !e.target.value ||
                (e.target.value && e.target.value.indexOf(PREFIX)) !== 0
            ) {
                e.target.value = PREFIX;
            }
        }
    }
}
