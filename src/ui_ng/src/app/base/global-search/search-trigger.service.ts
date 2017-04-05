import { Injectable } from '@angular/core';
import { Subject } from 'rxjs/Subject';
import { AlertType } from '../../shared/shared.const';

@Injectable()
export class SearchTriggerService {

  private searchTriggerSource = new Subject<string>();
  private searchCloseSource = new Subject<boolean>();
  private searchClearSource = new Subject<boolean>();

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
  clear(event): void {
    this.searchClearSource.next(event);
  }

}