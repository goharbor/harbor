
import { debounceTime, distinctUntilChanged } from 'rxjs/operators';
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
import { Component, OnInit, OnDestroy } from '@angular/core';
import { Router } from '@angular/router';
import { Subject ,  Subscription } from "rxjs";

import { SearchTriggerService } from './search-trigger.service';

import { AppConfigService } from '../../app-config.service';



import {TranslateService} from "@ngx-translate/core";
import {SkinableConfig} from "../../skinable-config.service";

const deBounceTime = 500; // ms

@Component({
    selector: 'global-search',
    templateUrl: "global-search.component.html",
    styleUrls: ["search.component.scss"]
})
export class GlobalSearchComponent implements OnInit, OnDestroy {
    // Keep search term as Subject
    searchTerms = new Subject<string>();

    // Keep subscription for future use
    searchSub: Subscription;
    closeSub: Subscription;

    // To indicate if the result panel is opened
    isResPanelOpened: boolean = false;
    searchTerm: string = "";

    placeholderText: string;

    constructor(
        private searchTrigger: SearchTriggerService,
        private router: Router,
        private appConfigService: AppConfigService,
        private translate: TranslateService,
        private skinableConfig: SkinableConfig) {
    }

    ngOnInit(): void {
        // custom skin
        let customSkinObj = this.skinableConfig.getProject();
        if (customSkinObj && customSkinObj.name) {
            this.translate.get('GLOBAL_SEARCH.PLACEHOLDER', {'param': customSkinObj.name}).subscribe(res => {
                // Placeholder text
                this.placeholderText = res;
            });
        } else {
            this.translate.get('GLOBAL_SEARCH.PLACEHOLDER', {'param': 'Harbor'}).subscribe(res => {
                // Placeholder text
                this.placeholderText = res;
            });
        }

        this.searchSub = this.searchTerms.pipe(
            debounceTime(deBounceTime),
            distinctUntilChanged())
            .subscribe(term => {
                this.searchTrigger.triggerSearch(term);
            });
        this.closeSub = this.searchTrigger.searchClearChan$.subscribe(clear => {
            this.searchTerm = "";
        });

        if (this.appConfigService.isIntegrationMode()) {
            this.placeholderText = "GLOBAL_SEARCH.PLACEHOLDER_VIC";
        }
    }

    ngOnDestroy(): void {
        if (this.searchSub) {
            this.searchSub.unsubscribe();
        }

        if (this.closeSub) {
            this.closeSub.unsubscribe();
        }
    }

    // Handle the term inputting event
    search(term: string): void {
        // Send event even term is empty

        this.searchTerms.next(term.trim());
    }
}
