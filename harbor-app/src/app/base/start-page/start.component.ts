import { Component, OnInit } from '@angular/core';

import { SessionService } from '../../shared/session.service';
import { SessionUser } from '../../shared/session-user';

@Component({
    selector: 'start-page',
    templateUrl: "start.component.html",
    styleUrls: ['start.component.css']
})
export class StartPageComponent implements OnInit {
    private currentUser: SessionUser = null;

    constructor(
        private session: SessionService
    ) { }

    public get currentUsername(): string {
        return this.currentUser ? this.currentUser.username : "";
    }

    //Implement ngOnIni
    ngOnInit(): void {
        this.currentUser = this.session.getCurrentUser();
    }
}