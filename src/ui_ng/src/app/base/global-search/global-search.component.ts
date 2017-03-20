import { Component, Output, EventEmitter, OnInit, OnDestroy } from '@angular/core';
import { Router } from '@angular/router';
import { Subject } from 'rxjs/Subject';
import { Observable } from 'rxjs/Observable';
import { Subscription } from 'rxjs/Subscription';

import { SearchTriggerService } from './search-trigger.service';

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
    private stateSub: Subscription;

    //To indicate if the result panel is opened
    private isResPanelOpened: boolean = false;

    constructor(
        private searchTrigger: SearchTriggerService,
        private router: Router) { }

    //Implement ngOnIni
    ngOnInit(): void {
        this.searchSub = this.searchTerms
            .debounceTime(deBounceTime)
            //.distinctUntilChanged()
            .subscribe(term => {
                this.searchTrigger.triggerSearch(term);
            });
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