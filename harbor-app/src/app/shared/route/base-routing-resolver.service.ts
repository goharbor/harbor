import { Injectable } from '@angular/core';
import {
    Router,
    Resolve,
    ActivatedRouteSnapshot,
    RouterStateSnapshot,
    NavigationExtras
} from '@angular/router';

import { SessionService } from '../../shared/session.service';
import { SessionUser } from '../../shared/session-user';
import { harborRootRoute } from '../shared.const';

@Injectable()
export class BaseRoutingResolver implements Resolve<SessionUser> {

    constructor(private session: SessionService, private router: Router) { }

    resolve(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Promise<SessionUser> {
        //To refresh seesion
        return this.session.retrieveUser()
            .then(sessionUser => {
                return sessionUser;
            })
            .catch(error => {
                //Session retrieving failed then redirect to sign-in
                //no matter what status code is.
                //Please pay attention that route 'harborRootRoute' support anonymous user
                if (state.url != harborRootRoute) {
                    let navigatorExtra: NavigationExtras = {
                        queryParams: { "redirect_url": state.url }
                    };
                    this.router.navigate(['sign-in'], navigatorExtra);
                }
            });
    }
}