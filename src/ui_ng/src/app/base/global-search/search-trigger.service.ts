import { Injectable } from '@angular/core';
import { Subject } from 'rxjs/Subject';
import { AlertType } from '../../shared/shared.const';

@Injectable()
export class SearchTriggerService {

  private searchTriggerSource = new Subject<string>();
  private searchCloseSource = new Subject<boolean>();
  private searchInputSource = new Subject<boolean>();

  searchTriggerChan$ = this.searchTriggerSource.asObservable();
  searchCloseChan$ = this.searchCloseSource.asObservable();
  searchInputChan$ = this.searchInputSource.asObservable();

  triggerSearch(event: string) {
    this.searchTriggerSource.next(event);
  }

  //Set event to true for shell
  //set to false for search panel
  closeSearch(event: boolean) {
    this.searchCloseSource.next(event);
  }

  //Notify the state change of search box in home start page
  searchInputStat(event: boolean) {
    this.searchInputSource.next(event);
  }

}