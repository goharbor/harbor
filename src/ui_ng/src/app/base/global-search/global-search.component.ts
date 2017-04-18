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
import { Component, Output, EventEmitter, OnInit, OnDestroy } from '@angular/core';
import { Router } from '@angular/router';
import { Subject } from 'rxjs/Subject';
import { Observable } from 'rxjs/Observable';
import { Subscription } from 'rxjs/Subscription';

import { SearchTriggerService } from './search-trigger.service';

import { AppConfigService } from '../../app-config.service';

import 'rxjs/add/operator/debounceTime';
import 'rxjs/add/operator/distinctUntilChanged';

const deBounceTime = 500; //ms

@Component({
    selector: 'global-search',
    templateUrl: "global-search.component.html"
})
export class GlobalSearchComponent implements OnInit, OnDestroy {
    //Keep search term as Subject
    private searchTerms = new Subject<string>();

    //Keep subscription for future use
    private searchSub: Subscription;
    private closeSub: Subscription;

    //To indicate if the result panel is opened
    private isResPanelOpened: boolean = false;
    private searchTerm: string = "";

    //Placeholder text
    placeholderText: string = "GLOBAL_SEARCH.PLACEHOLDER";

    constructor(
        private searchTrigger: SearchTriggerService,
        private router: Router,
        private appConfigService: AppConfigService) { }

    //Implement ngOnIni
    ngOnInit(): void {
        this.searchSub = this.searchTerms
            .debounceTime(deBounceTime)
            //.distinctUntilChanged()
            .subscribe(term => {
                this.searchTrigger.triggerSearch(term);
            });
        this.closeSub = this.searchTrigger.searchClearChan$.subscribe(clear => {
            this.searchTerm = "";
        });

        if(this.appConfigService.isIntegrationMode()){
            this.placeholderText = "GLOBAL_SEARCH.PLACEHOLDER_VIC";
        }
    }

    ngOnDestroy(): void {
        if (this.searchSub) {
            this.searchSub.unsubscribe();
        }
    }

    //Handle the term inputting event
    search(term: string): void {
        //Send event even term is empty

        this.searchTerms.next(term.trim());
    }
}