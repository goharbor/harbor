import { Component, Output, EventEmitter } from '@angular/core';

import { GlobalSearchService } from './global-search.service';
import { SearchResults } from './search-results';

@Component({
    selector: "search-result",
    templateUrl: "search-result.component.html",
    styleUrls: ["search-result.component.css"],

    providers: [GlobalSearchService]
})

export class SearchResultComponent {
    @Output() closeEvt = new EventEmitter<boolean>();

    searchResults: SearchResults;

    //Open or close
    private stateIndicator: boolean = false;
    //Search in progress
    private onGoing: boolean = true;

    //Whether or not mouse point is onto the close indicator
    private mouseOn: boolean = false;

    constructor(private search: GlobalSearchService) { }

    public get state(): boolean {
        return this.stateIndicator;
    }

    public get done(): boolean {
        return !this.onGoing;
    }

    public get hover(): boolean {
        return this.mouseOn;
    }

    //Handle mouse event of close indicator
    mouseAction(over: boolean): void {
        this.mouseOn = over;
    }
    
    //Show the results
    show(): void {
        this.stateIndicator = true;
    }

    //Close the result page
    close(): void {
        //Tell shell close
        this.closeEvt.emit(true);

        this.stateIndicator = false;
    }

    //Call search service to complete the search request
    doSearch(term: string): void {
        //Confirm page is displayed
        if (!this.stateIndicator) {
            this.show();
        }

        //Show spinner
        this.onGoing = true;

        this.search.doSearch(term)
            .then(searchResults => {
                this.onGoing = false;
                this.searchResults = searchResults;
                console.info(searchResults);
            })
            .catch(error => {
                this.onGoing = false;
                console.error(error);//TODO: Use general erro handler
            });
    }
}