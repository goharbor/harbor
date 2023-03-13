import { Component, OnInit, OnDestroy, HostListener } from '@angular/core';
import {
    downloadCVEs,
    EventState,
    ExportJobStatus,
    OperationService,
} from './operation.service';
import { forkJoin, Subscription } from 'rxjs';
import {
    OperateInfo,
    OperateInfosLocalstorage,
    OperationState,
} from './operate';
import { SlideInOutAnimation } from '../../_animations/slide-in-out.animation';
import { TranslateService } from '@ngx-translate/core';
import { SessionService } from '../../services/session.service';
import { ScanDataExportService } from '../../../../../ng-swagger-gen/services/scan-data-export.service';
import {
    EventService,
    HarborEvent,
} from '../../../services/event-service/event.service';
import { MessageHandlerService } from '../../services/message-handler.service';
import { HarborDatetimePipe } from '../../pipes/harbor-datetime.pipe';
const STAY_TIME: number = 5000;
const OPERATION_KEY: string = 'operation';
const MAX_NUMBER: number = 500;
const MAX_SAVING_TIME: number = 1000 * 60 * 60 * 24 * 30; // 30 days
const TIMEOUT = 7000;
const FILE_NAME_PREFIX: string = 'csv_file_';
const RETRY_TIMES: number = 50;
@Component({
    selector: 'hbr-operation-model',
    templateUrl: './operation.component.html',
    styleUrls: ['./operation.component.css'],
    animations: [SlideInOutAnimation],
})
export class OperationComponent implements OnInit, OnDestroy {
    batchInfoSubscription: Subscription;
    resultLists: OperateInfo[] = [];
    exportJobs: OperateInfo[] = [];
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
    timeout;
    refreshExportJobSub: Subscription;
    retryTimes: number = RETRY_TIMES;
    constructor(
        private session: SessionService,
        private operationService: OperationService,
        private translate: TranslateService,
        private scanDataExportService: ScanDataExportService,
        private event: EventService,
        private msgHandler: MessageHandlerService
    ) {
        if (!this.refreshExportJobSub) {
            this.refreshExportJobSub = this.event.subscribe(
                HarborEvent.REFRESH_EXPORT_JOBS,
                () => {
                    if (this.animationState === 'out') {
                        this._newMessageCount += 1;
                    }
                    this.refreshExportJobs(false);
                }
            );
        }
        if (!this.batchInfoSubscription) {
            this.batchInfoSubscription =
                operationService.operationInfo$.subscribe(data => {
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
                });
        }
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
            }, STAY_TIME);
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
            this.refreshExportJobs(false);
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
            this.batchInfoSubscription = null;
        }
        if (this._timeoutInterval) {
            clearInterval(this._timeoutInterval);
            this._timeoutInterval = null;
        }
        if (this.timeout) {
            clearTimeout(this.timeout);
            this.timeout = null;
        }
        if (this.refreshExportJobSub) {
            this.refreshExportJobSub.unsubscribe();
            this.refreshExportJobSub = null;
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
        this.refreshExportJobs(false);
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
    refreshExportJobs(isRetry: boolean) {
        if (this.session.getCurrentUser()) {
            if (isRetry) {
                this.retryTimes--;
            } else {
                this.retryTimes = RETRY_TIMES;
            }
            this.scanDataExportService
                .getScanDataExportExecutionList()
                .subscribe(res => {
                    if (res?.items) {
                        this.exportJobs = [];
                        let flag: boolean = false;
                        res.items.forEach(item => {
                            const info: OperateInfo = {
                                name: 'CVE_EXPORT.EXPORT_TITLE',
                                state: this.MapStatus(item.status),
                                data: {
                                    hasFile: item.file_present,
                                    name: `${FILE_NAME_PREFIX}${new HarborDatetimePipe().transform(
                                        item.start_time,
                                        'yyyyMMddHHmmss'
                                    )}`,
                                    id: item.id,
                                    errorInf: item.status_text
                                        ? item.status_text
                                        : null,
                                },
                                timeStamp: new Date(item.start_time).getTime(),
                                timeDiff: 'OPERATION.SECOND_AGO',
                            };
                            this.exportJobs.push(info);
                            if (this.isRunningState(item.status)) {
                                flag = true;
                            }
                        });
                        this.refreshTimestampForExportJob();
                        if (flag && this.retryTimes > 0) {
                            if (this.timeout) {
                                clearTimeout(this.timeout);
                                this.timeout = null;
                            }
                            this.timeout = setTimeout(() => {
                                this.refreshExportJobs(true);
                            }, TIMEOUT);
                        }
                    }
                });
        }
    }

    isRunningState(state: string): boolean {
        if (state) {
            return (
                state === ExportJobStatus.RUNNING ||
                state === ExportJobStatus.PENDING ||
                state === ExportJobStatus.SCHEDULED
            );
        }
        return false;
    }
    MapStatus(originStatus: string): string {
        if (originStatus) {
            if (this.isRunningState(originStatus)) {
                return EventState.PROGRESSING;
            }
            if (originStatus === ExportJobStatus.STOPPED) {
                return EventState.INTERRUPT;
            }
            if (originStatus === ExportJobStatus.SUCCESS) {
                return EventState.SUCCESS;
            }
            if (originStatus === ExportJobStatus.ERROR) {
                return EventState.FAILURE;
            }
        }
        return EventState.FAILURE;
    }
    download(info: OperateInfo) {
        if (info?.data?.id && info?.data?.name) {
            this.scanDataExportService
                .downloadScanData({
                    executionId: +info.data.id,
                })
                .subscribe(
                    res => {
                        downloadCVEs(res, info.data.name);
                        this.refreshExportJobs(false);
                    },
                    error => {
                        this.msgHandler.error(error);
                    }
                );
        }
    }
    refreshTimestampForExportJob() {
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
        if (this.exportJobs?.length) {
            this.exportJobs.forEach(data => {
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
    }
}
