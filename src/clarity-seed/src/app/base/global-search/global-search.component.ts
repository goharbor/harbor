import { Component, Output, EventEmitter, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { Subject } from 'rxjs/Subject';
import { Observable } from 'rxjs/Observable';
import { SearchEvent } from '../search-event';

import 'rxjs/add/operator/debounceTime';
import 'rxjs/add/operator/distinctUntilChanged';

const deBounceTime = 500; //ms

@Component({
    selector: 'global-search',
    templateUrl: "global-search.component.html"
})
export class GlobalSearchComponent implements OnInit {
    //Publish search event to parent
    @Output() searchEvt = new EventEmitter<SearchEvent>();

    //Keep search term as Subject
    private searchTerms = new Subject<string>();

    //Implement ngOnIni
    ngOnInit(): void {
        this.searchTerms
            .debounceTime(deBounceTime)
            .distinctUntilChanged()
            .subscribe(term => {
                this.searchEvt.emit({
                    term: term
                });
            });
    }

    //Handle the term inputting event
    search(term: string): void {
        //Send event only when term is not empty

        let nextTerm = term.trim();
        if (nextTerm != "") {
            this.searchTerms.next(nextTerm);
        }
    }
}