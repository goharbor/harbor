import { Injectable } from '@angular/core';
import { Subject } from 'rxjs/Subject';
import { AlertType } from '../../shared/shared.const';

@Injectable()
export class SearchTriggerService {

  searchTriggerSource = new Subject<string>();
  searchCloseSource = new Subject<boolean>();
  searchClearSource = new Subject<boolean>();

  searchTriggerChan$ = this.searchTriggerSource.asObservable();
  searchCloseChan$ = this.searchCloseSource.asObservable();
  searchClearChan$ = this.searchClearSource.asObservable();

  triggerSearch(event: string) {
    this.searchTriggerSource.next(event);
  }

  //Set event to true for shell
  //set to false for search panel
  closeSearch(event: boolean) {
    this.searchCloseSource.next(event);
  }

  //Clear search term
  clear(event: any): void {
    this.searchClearSource.next(event);
  }

}