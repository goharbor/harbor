import { Component, OnInit, OnDestroy, HostListener } from '@angular/core';
import { OperationService } from './operation.service';
import { forkJoin, Subscription } from 'rxjs';
import {
    OperateInfo,
    OperateInfosLocalstorage,
    OperationState,
} from './operate';
import { SlideInOutAnimation } from '../../_animations/slide-in-out.animation';
import { TranslateService } from '@ngx-translate/core';
import { SessionService } from '../../services/session.service';

const OPERATION_KEY: string = 'operation';
const MAX_NUMBER: number = 500;
const MAX_SAVING_TIME: number = 1000 * 60 * 60 * 24 * 30; // 30 days

@Component({
    selector: 'hbr-operation-model',
    templateUrl: './operation.component.html',
    styleUrls: ['./operation.component.css'],
    animations: [SlideInOutAnimation],
})
export class OperationComponent implements OnInit, OnDestroy {
    batchInfoSubscription: Subscription;
    resultLists: OperateInfo[] = [];
    animationState = 'out';
    private _newMessageCount: number = 0;
    private _timeoutInterval;

    @HostListener('window:beforeunload', ['$event'])
    beforeUnloadHander(event) {
        if (this.session.getCurrentUser()) {
            // store into localStorage
            // group by user id
            localStorage.setItem(
                `${OPERATION_KEY}-${this.session.getCurrentUser().user_id}`,
                JSON.stringify({
                    updated: new Date().getTime(),
                    data: this.resultLists,
                    newMessageCount: this._newMessageCount,
                })
            );
        }
    }

    constructor(
        private session: SessionService,
        private operationService: OperationService,
        private translate: TranslateService
    ) {
        this.batchInfoSubscription = operationService.operationInfo$.subscribe(
            data => {
                if (this.animationState === 'out') {
                    this._newMessageCount += 1;
                }
                if (data) {
                    if (this.resultLists.length >= MAX_NUMBER) {
                        this.resultLists.splice(
                            MAX_NUMBER - 1,
                            this.resultLists.length + 1 - MAX_NUMBER
                        );
                    }
                    this.resultLists.unshift(data);
                }
            }
        );
    }

    getNewMessageCountStr(): string {
        if (this._newMessageCount) {
            if (this._newMessageCount > MAX_NUMBER) {
                return MAX_NUMBER + '+';
            }
            return this._newMessageCount.toString();
        }
        return '';
    }

    resetNewMessageCount() {
        this._newMessageCount = 0;
    }

    mouseover() {
        if (this._timeoutInterval) {
            clearInterval(this._timeoutInterval);
            this._timeoutInterval = null;
        }
    }

    mouseleave() {
        if (!this._timeoutInterval) {
            this._timeoutInterval = setTimeout(() => {
                this.animationState = 'out';
            }, 5000);
        }
    }

    public get runningLists(): OperateInfo[] {
        let runningList: OperateInfo[] = [];
        this.resultLists.forEach(data => {
            if (data.state === 'progressing') {
                runningList.push(data);
            }
        });
        return runningList;
    }

    public get failLists(): OperateInfo[] {
        let failedList: OperateInfo[] = [];
        this.resultLists.forEach(data => {
            if (data.state === 'failure') {
                failedList.push(data);
            }
        });
        return failedList;
    }

    init() {
        if (this.session.getCurrentUser()) {
            const operationInfosString: string = localStorage.getItem(
                `${OPERATION_KEY}-${this.session.getCurrentUser().user_id}`
            );
            if (operationInfosString) {
                const operationInfos: OperateInfosLocalstorage =
                    JSON.parse(operationInfosString);
                if (operationInfos) {
                    if (operationInfos.newMessageCount) {
                        this._newMessageCount = operationInfos.newMessageCount;
                    }
                    if (operationInfos.data && operationInfos.data.length) {
                        // remove expired items
                        operationInfos.data = operationInfos.data.filter(
                            item => {
                                return (
                                    new Date().getTime() - item.timeStamp <
                                    MAX_SAVING_TIME
                                );
                            }
                        );
                        operationInfos.data.forEach(operInfo => {
                            if (operInfo.state === OperationState.progressing) {
                                operInfo.state = OperationState.interrupt;
                                operInfo.data.errorInf =
                                    'operation been interrupted';
                            }
                        });
                        this.resultLists = operationInfos.data;
                    }
                }
            }
        }
    }

    ngOnInit() {
        this.init();
    }

    ngOnDestroy(): void {
        if (this.batchInfoSubscription) {
            this.batchInfoSubscription.unsubscribe();
        }
        if (this._timeoutInterval) {
            clearInterval(this._timeoutInterval);
            this._timeoutInterval = null;
        }
    }

    toggleTitle(errorSpan: any) {
        errorSpan.style.display =
            errorSpan.style.display === 'block' ? 'none' : 'block';
    }

    slideOut(): void {
        this.animationState = this.animationState === 'out' ? 'in' : 'out';
        if (this.animationState === 'in') {
            this.resetNewMessageCount();
            // refresh when open
            this.TabEvent();
        }
    }

    openSlide(): void {
        this.animationState = 'in';
        this.resetNewMessageCount();
    }

    TabEvent(): void {
        let secondsAgo: string,
            minutesAgo: string,
            hoursAgo: string,
            daysAgo: string;
        forkJoin([
            this.translate.get('OPERATION.SECOND_AGO'),
            this.translate.get('OPERATION.MINUTE_AGO'),
            this.translate.get('OPERATION.HOUR_AGO'),
            this.translate.get('OPERATION.DAY_AGO'),
        ]).subscribe(res => {
            [secondsAgo, minutesAgo, hoursAgo, daysAgo] = res;
        });
        this.resultLists.forEach(data => {
            const timeDiff: number = new Date().getTime() - +data.timeStamp;
            data.timeDiff = this.calculateTime(
                timeDiff,
                secondsAgo,
                minutesAgo,
                hoursAgo,
                daysAgo
            );
        });
    }

    calculateTime(
        timeDiff: number,
        s: string,
        m: string,
        h: string,
        d: string
    ) {
        const dist = Math.floor(timeDiff / 1000 / 60); // change to minute;
        if (dist > 0 && dist < 60) {
            return Math.floor(dist) + m;
        } else if (dist >= 60 && Math.floor(dist / 60) < 24) {
            return Math.floor(dist / 60) + h;
        } else if (Math.floor(dist / 60) >= 24) {
            return Math.floor(dist / 60 / 24) + d;
        } else {
            return s;
        }
    }
}
