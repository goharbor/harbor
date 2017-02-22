import { Injectable } from '@angular/core';
import {
    Router, Resolve, ActivatedRouteSnapshot, RouterStateSnapshot
} from '@angular/router';

import { SessionService } from '../shared/session.service';
import { SessionUser } from '../shared/session-user';

@Injectable()
export class BaseRoutingResolver implements Resolve<SessionUser> {

    constructor(private session: SessionService, private router: Router) { }

    resolve(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Promise<SessionUser> {
        return this.session.retrieveUser()
            .then(sessionUser => {
                return sessionUser;
            })
            .catch(error => {
                console.info("Anonymous user");
            });
    }
}