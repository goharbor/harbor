import { Component, Output, EventEmitter, OnInit, OnDestroy } from '@angular/core';
import { Router } from '@angular/router';
import { Subject } from 'rxjs/Subject';
import { Observable } from 'rxjs/Observable';;
import { Subscription } from 'rxjs/Subscription';

import { SessionService } from '../../shared/session.service';
import { SessionUser } from '../../shared/session-user';

import { SearchTriggerService } from './search-trigger.service';

import 'rxjs/add/operator/debounceTime';
import 'rxjs/add/operator/distinctUntilChanged';

const deBounceTime = 500; //ms

@Component({
    selector: 'search-start',
    templateUrl: "search-start.component.html",
    styleUrls: ['search-start.component.css']
})
export class SearchStartComponent implements OnInit, OnDestroy {
    //Keep search term as Subject
    private searchTerms = new Subject<string>();

    private searchSub: Subscription;

    private currentUser: SessionUser = null;

    constructor(
        private session: SessionService,
        private searchTrigger: SearchTriggerService){}

    public get currentUsername(): string {
        return this.currentUser?this.currentUser.username: "";
    }

    //Implement ngOnIni
    ngOnInit(): void {
        this.currentUser = this.session.getCurrentUser();

        this.searchSub = this.searchTerms
            .debounceTime(deBounceTime)
            .distinctUntilChanged()
            .subscribe(term => {
                this.searchTrigger.triggerSearch(term);
            });
    }

    ngOnDestroy(): void {
        if(this.searchSub){
            this.searchSub.unsubscribe();
        }
    }

    //Handle the term inputting event
    search(term: string): void {
        //Send event only when term is not empty

        this.searchTerms.next(term);
    }
}