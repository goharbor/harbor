import { Injectable } from '@angular/core';
import { Headers, Http } from '@angular/http';
import 'rxjs/add/operator/toPromise';

import { SessionUser } from './session-user';

const currentUserEndpint = "/api/users/current";
const signOffEndpoint = "/log_out";
/**
 * Define related methods to handle account and session corresponding things
 * 
 * @export
 * @class SessionService
 */
@Injectable()
export class SessionService {
    currentUser: SessionUser = null;

    private headers = new Headers({
        "Content-Type": 'application/json'
    });

    constructor(private http: Http) {}

    /**
     * Get the related information of current signed in user from backend
     * 
     * @returns {Promise<SessionUser>}
     * 
     * @memberOf SessionService
     */
    retrieveUser(): Promise<SessionUser> {
        return this.http.get(currentUserEndpint, { headers: this.headers }).toPromise()
            .then(response => {
                this.currentUser = response.json() as SessionUser;
                return this.currentUser;
            })
            .catch(error => {
                console.log("An error occurred when getting current user ", error);//TODO
                return Promise.reject(error);
            })
    }

    /**
     * For getting info
     */
    getCurrentUser(): SessionUser {
        return this.currentUser;
    }

    /**
     * Log out the system
     */
    signOff(): Promise<any> {
        return this.http.get(signOffEndpoint, { headers: this.headers }).toPromise()
        .then(() => {
            //Destroy current session cache
            this.currentUser = null;
        }) //Nothing returned
        .catch(error => {
            console.log("An error occurred when signing off ", error);//TODO
            return Promise.reject(error);
        })
    }
}