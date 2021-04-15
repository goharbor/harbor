import { Component, OnInit, OnDestroy, HostListener } from '@angular/core';
import { OperationService } from "./operation.service";
import { Subscription } from "rxjs";
import { OperateInfo, OperationState } from "./operate";
import { SlideInOutAnimation } from "../../_animations/slide-in-out.animation";
import { TranslateService } from "@ngx-translate/core";
import { SessionService } from "../../services/session.service";

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
  animationState = "out";
  private _newMessageCount: number = 0;
  private _timeoutInterval;

  @HostListener('window:beforeunload', ['$event'])
  beforeUnloadHander(event) {
    if (this.session.getCurrentUser()) {
      // storage to localStorage
      const timp = new Date().getTime();
      // group by user id
      localStorage.setItem(`${OPERATION_KEY}-${this.session.getCurrentUser().user_id}`,
        JSON.stringify({
          timp: timp,
          data: this.resultLists,
          newMessageCount: this._newMessageCount
        }));
    }
  }

  constructor(
    private session: SessionService,
    private operationService: OperationService,
    private translate: TranslateService) {

    this.batchInfoSubscription = operationService.operationInfo$.subscribe(data => {
      if (this.animationState === 'out') {
        this._newMessageCount += 1;
      }
      if (data) {
        if (this.resultLists.length >= MAX_NUMBER) {
          this.resultLists.splice(MAX_NUMBER - 1, this.resultLists.length + 1 - MAX_NUMBER);
        }
        this.resultLists.unshift(data);
      }
    });
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
      let requestCookie = localStorage.getItem(`${OPERATION_KEY}-${this.session.getCurrentUser().user_id}`);
      if (requestCookie) {
        let operInfors: any = JSON.parse(requestCookie);
        if (operInfors) {
          if (operInfors.newMessageCount) {
            this._newMessageCount = operInfors.newMessageCount;
          }
          if ((new Date().getTime() - operInfors.timp) > MAX_SAVING_TIME) {
            localStorage.removeItem(`${OPERATION_KEY}-${this.session.getCurrentUser().user_id}`);
          } else {
            if (operInfors.data) {
              operInfors.data.forEach(operInfo => {
                if (operInfo.state === OperationState.progressing) {
                  operInfo.state = OperationState.interrupt;
                  operInfo.data.errorInf = 'operation been interrupted';
                }
              });
              this.resultLists = operInfors.data;
            }
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
    errorSpan.style.display = (errorSpan.style.display === 'block') ? 'none' : 'block';
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
    let timp: any;
    this.resultLists.forEach(data => {
      timp = new Date().getTime() - +data.timeStamp;
      data.timeDiff = this.calculateTime(timp);
    });
  }

  calculateTime(timp: number) {
    let dist = Math.floor(timp / 1000 / 60);  // change to minute;
    if (dist > 0 && dist < 60) {
      return Math.floor(dist) + ' minute(s) ago';
    } else if (dist >= 60 && Math.floor(dist / 60) < 24) {
      return Math.floor(dist / 60) + ' hour(s) ago';
    } else if (Math.floor(dist / 60) >= 24) {
      return Math.floor(dist / 60 / 24) + ' day(s) ago';
    } else {
      return 'less than 1 minute';
    }
  }

  translateTime(tim: string, param?: number) {
    this.translate.get(tim, {'param': param}).subscribe((res: string) => {
      return res;
    });
  }
}
