import { Injectable } from '@angular/core';
import { Headers, Http, RequestOptions } from '@angular/http';
import 'rxjs/add/operator/toPromise';

import { User } from './user';

const userMgmtEndpoint = '/api/users';

/**
 * Define related methods to handle account and session corresponding things
 * 
 * @export
 * @class SessionService
 */
@Injectable()
export class UserService {
    private httpOptions = new RequestOptions({
        headers: new Headers({
            "Content-Type": 'application/json'
        })
    });

    constructor(private http: Http) { }

    //Handle the related exceptions
    private handleError(error: any): Promise<any> {
        return Promise.reject(error.message || error);
    }

    //Get the user list
    getUsers(): Promise<User[]> {
        return this.http.get(userMgmtEndpoint, this.httpOptions).toPromise()
            .then(response => response.json() as User[])
            .catch(error => this.handleError(error));
    }

    //Add new user
    addUser(user: User): Promise<any> {
        return this.http.post(userMgmtEndpoint, JSON.stringify(user), this.httpOptions).toPromise()
            .then(() => null)
            .catch(error => this.handleError(error));
    }

    //Delete the specified user
    deleteUser(userId: number): Promise<any> {
        return this.http.delete(userMgmtEndpoint + "/" + userId, this.httpOptions)
            .toPromise()
            .then(() => null)
            .catch(error => this.handleError(error));
    }

    //Update user to enable/disable the admin role
    updateUser(user: User): Promise<any> {
        return this.http.put(userMgmtEndpoint + "/" + user.user_id, JSON.stringify(user), this.httpOptions)
            .toPromise()
            .then(() => null)
            .catch(error => this.handleError(error));
    }

    //Set user admin role
    updateUserRole(user: User): Promise<any> {
        return this.http.put(userMgmtEndpoint + "/" + user.user_id + "/sysadmin", JSON.stringify(user), this.httpOptions)
            .toPromise()
            .then(() => null)
            .catch(error => this.handleError(error));
    }
}