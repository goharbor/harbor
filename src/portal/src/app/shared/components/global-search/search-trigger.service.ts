// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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
import { Injectable } from '@angular/core';
import { Subject } from 'rxjs';

@Injectable({
    providedIn: 'root',
})
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

    // Set event to true for shell
    // set to false for search panel
    closeSearch(event: boolean) {
        this.searchCloseSource.next(event);
    }

    // Clear search term
    clear(event: any): void {
        this.searchClearSource.next(event);
    }
}
