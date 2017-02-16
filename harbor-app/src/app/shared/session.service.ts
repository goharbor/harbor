import { Injectable } from '@angular/core';
import { Headers, Http } from '@angular/http';
import 'rxjs/add/operator/toPromise';

const currentUserEndpint = "/api/users/current";
/**
 * Define related methods to handle account and session corresponding things
 * 
 * @export
 * @class SessionService
 */
@Injectable()
export class SessionService {
    currentUser: any = null;

    private headers = new Headers({
        "Content-Type": 'application/json'
    });

    constructor(private http: Http) {}

    /**
     * Get the related information of current signed in user from backend
     * 
     * @returns {Promise<any>}
     * 
     * @memberOf SessionService
     */
    retrieveUser(): Promise<any> {
        return this.http.get(currentUserEndpint, { headers: this.headers }).toPromise()
            .then(response => this.currentUser = response.json())
            .catch(error => {
                console.log("An error occurred when getting current user ", error);//TODO: Will replaced with general error handler
            })
    }

    /**
     * For getting info
     */
    getCurrentUser(): any {
        return this.currentUser;
    }
}