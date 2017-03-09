import { Component, Output, EventEmitter, OnInit, OnDestroy } from '@angular/core';
import { Router } from '@angular/router';
import { Subject } from 'rxjs/Subject';
import { Observable } from 'rxjs/Observable';;
import { Subscription } from 'rxjs/Subscription';

import { SessionService } from '../../shared/session.service';
import { SessionUser } from '../../shared/session-user';

import { SearchTriggerService } from '../global-search/search-trigger.service';

import { Repository } from '../../repository/repository';
import { TopRepoService } from './top-repository.service';

import { errorHandler } from '../../shared/shared.utils';
import { AlertType } from '../../shared/shared.const';

import { MessageService } from '../../global-message/message.service';

import 'rxjs/add/operator/debounceTime';
import 'rxjs/add/operator/distinctUntilChanged';

const deBounceTime = 500; //ms

@Component({
    selector: 'start-page',
    templateUrl: "start.component.html",
    styleUrls: ['start.component.css'],

    providers: [TopRepoService]
})
export class StartPageComponent implements OnInit, OnDestroy {
    //Keep search term as Subject
    private searchTerms = new Subject<string>();

    private searchSub: Subscription;

    private currentUser: SessionUser = null;

    private topRepos: Repository[] = [];

    constructor(
        private session: SessionService,
        private searchTrigger: SearchTriggerService,
        private topRepoService: TopRepoService,
        private msgService: MessageService
    ) { }

    public get currentUsername(): string {
        return this.currentUser ? this.currentUser.username : "";
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

        this.getTopRepos();
    }

    ngOnDestroy(): void {
        if (this.searchSub) {
            this.searchSub.unsubscribe();
        }
    }

    //Handle the term inputting event
    search(term: string): void {
        //Send event only when term is not empty

        this.searchTerms.next(term);
    }

    //Get top popular repositories
    getTopRepos() {
        this.topRepoService.getTopRepos()
            .then(repos => repos.forEach(item => {
                this.topRepos.push(new Repository(item.name, item.count));
            }))
            .catch(error => {
                this.msgService.announceMessage(error.status, errorHandler(error), AlertType.WARNING);
            })
    }
}