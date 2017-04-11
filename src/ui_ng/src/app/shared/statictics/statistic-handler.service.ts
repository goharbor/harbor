import { Injectable } from '@angular/core';
import { Subject } from 'rxjs/Subject';

@Injectable()
export class StatisticHandler {

  refreshSource = new Subject<boolean>();

  refreshChan$ = this.refreshSource.asObservable();

  refresh() {
    this.refreshSource.next(true);
  }
}