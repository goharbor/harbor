// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
import {
    Component,
    Input,
    OnInit,
    OnDestroy,
    SimpleChanges,
    OnChanges,
} from '@angular/core';
import { RobotTimeRemainColor } from '../../../base/left-side-nav/system-robot-accounts/system-robot-util';
const SEC: number = 1000;
const MIN: number = 60 * 1000;
const DAY: number = 1000 * 60 * 60 * 24;
const HOUR: number = 1000 * 60 * 60;
const WARNING_DAYS = 7;
@Component({
    selector: 'app-remaining-time',
    templateUrl: 'remaining-time.component.html',
    styleUrls: ['./remaining-time.component.scss'],
})
export class RemainingTimeComponent implements OnInit, OnDestroy, OnChanges {
    color: string;
    timeRemain: string;
    @Input()
    timeDiff: number; // the different between server time and local time, unit millisecond, localTime - serverTime
    @Input()
    deadline: number; // unit second
    intelVal: any;
    constructor() {}

    ngOnInit() {
        if (!this.intelVal) {
            this.intelVal = setInterval(() => {
                this.refreshTimeRemain();
            }, MIN);
        }
    }
    ngOnDestroy() {
        if (this.intelVal) {
            clearInterval(this.intelVal);
            this.intelVal = null;
        }
    }
    ngOnChanges(changes: SimpleChanges): void {
        if (changes && changes['timeDiff'] && changes['deadline']) {
            this.refreshTimeRemain();
        }
    }
    refreshTimeRemain() {
        if (this.timeDiff !== null && this.deadline !== null) {
            if (this.deadline === -1) {
                this.color = RobotTimeRemainColor.GREEN;
                this.timeRemain = 'ROBOT_ACCOUNT.NEVER_EXPIRED';
                return;
            }
            const time =
                new Date(this.deadline * SEC).getTime() -
                new Date(new Date().getTime() - this.timeDiff).getTime();
            if (time > 0) {
                const days = Math.floor(time / DAY);
                const hours = Math.floor((time % DAY) / HOUR);
                const minutes = Math.floor((time % HOUR) / MIN);
                this.timeRemain = `${days}d ${hours}h ${minutes}m`;
                if (days >= WARNING_DAYS) {
                    this.color = RobotTimeRemainColor.GREEN;
                } else {
                    this.color = RobotTimeRemainColor.WARNING;
                }
            } else {
                this.color = RobotTimeRemainColor.EXPIRED;
                this.timeRemain = 'SYSTEM_ROBOT.EXPIRED';
            }
        }
    }
    isError(): boolean {
        return this.color === RobotTimeRemainColor.EXPIRED;
    }
    isWarning(): boolean {
        return this.color === RobotTimeRemainColor.WARNING;
    }
    isSuccess(): boolean {
        return this.color === RobotTimeRemainColor.GREEN;
    }
}
