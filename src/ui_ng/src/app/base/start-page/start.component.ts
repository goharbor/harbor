import { Component, OnInit } from '@angular/core';

import { SessionService } from '../../shared/session.service';
import { SessionUser } from '../../shared/session-user';

@Component({
    selector: 'start-page',
    templateUrl: "start.component.html",
    styleUrls: ['start.component.css']
})
export class StartPageComponent implements OnInit{
    private isSessionValid: boolean = false;

    constructor(
        private session: SessionService
    ) { }

    ngOnInit(): void {
        this.isSessionValid = this.session.getCurrentUser() != null;
    }
}