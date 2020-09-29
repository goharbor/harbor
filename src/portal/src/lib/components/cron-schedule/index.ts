import { Type } from "@angular/core";

import { CronScheduleComponent } from "./cron-schedule.component";
import { CronTooltipComponent } from "./cron-tooltip/cron-tooltip.component";

export * from "./cron-schedule.component";
export * from './cron-tooltip/cron-tooltip.component';
export const CRON_SCHEDULE_DIRECTIVES: Type<any>[] = [
    CronScheduleComponent,
    CronTooltipComponent
];
