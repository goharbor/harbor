import { Component, Output, EventEmitter, OnInit, OnDestroy } from '@angular/core';

import { GlobalSearchService } from './global-search.service';
import { SearchResults } from './search-results';
import { errorHandler, accessErrorHandler } from '../../shared/shared.utils';
import { AlertType, ListMode } from '../../shared/shared.const';
import { MessageService } from '../../global-message/message.service';

import { SearchTriggerService } from './search-trigger.service';

import { Subscription } from 'rxjs/Subscription';

@Component({
    selector: "search-result",
    templateUrl: "search-result.component.html",
    styleUrls: ["search-result.component.css"],

    providers: [GlobalSearchService]
})

export class SearchResultComponent implements OnInit, OnDestroy {
    private searchResults: SearchResults = new SearchResults();
    private originalCopy: SearchResults;

    private currentTerm: string = "";

    //Open or close
    private stateIndicator: boolean = false;
    //Search in progress
    private onGoing: boolean = false;

    //Whether or not mouse point is onto the close indicator
    private mouseOn: boolean = false;

    //Watch message channel
    private searchSub: Subscription;
    private closeSearchSub: Subscription;

    constructor(
        private search: GlobalSearchService,
        private msgService: MessageService,
        private searchTrigger: SearchTriggerService) { }

    ngOnInit() {
        this.searchSub = this.searchTrigger.searchTriggerChan$.subscribe(term => {
            this.doSearch(term);
        });
        this.closeSearchSub = this.searchTrigger.searchCloseChan$.subscribe(close => {
            this.close();
        });
    }

    ngOnDestroy() {
        if (this.searchSub) {
            this.searchSub.unsubscribe();
        }

        if (this.closeSearchSub) {
            this.closeSearchSub.unsubscribe();
        }
    }

    private clone(src: SearchResults): SearchResults {
        let res: SearchResults = new SearchResults();

        if (src) {
            src.project.forEach(pro => res.project.push(Object.assign({}, pro)));
            src.repository.forEach(repo => res.repository.push(Object.assign({}, repo)))

            return res;
        }

        return res//Empty object
    }

    public get listMode(): string {
        return ListMode.READONLY;
    }

    public get state(): boolean {
        return this.stateIndicator;
    }

    public get done(): boolean {
        return !this.onGoing;
    }

    public get hover(): boolean {
        return this.mouseOn;
    }

    //Show the results
    show(): void {
        this.stateIndicator = true;
    }

    //Close the result page
    close(): void {
        this.stateIndicator = false;
        this.searchTrigger.clear(true);
    }

    //Call search service to complete the search request
    doSearch(term: string): void {
        //Only search none empty term
        if (!term || term.trim() === "") {
            return;
        }
        //Do nothing if search is ongoing
        if (this.onGoing) {
            return;
        }
        //Confirm page is displayed
        if (!this.stateIndicator) {
            this.show();
        }

        this.currentTerm = term;

        //If term is empty, then clear the results
        if (term === "") {
            this.searchResults.project = [];
            this.searchResults.repository = [];
            return;
        }
        //Show spinner
        this.onGoing = true;

        this.search.doSearch(term)
            .then(searchResults => {
                this.onGoing = false;
                this.originalCopy = searchResults; //Keeo the original data
                this.searchResults = this.clone(searchResults);
            })
            .catch(error => {
                this.onGoing = false;
                if (!accessErrorHandler(error, this.msgService)) {
                    this.msgService.announceMessage(error.status, errorHandler(error), AlertType.DANGER);
                }
            });
    }
}