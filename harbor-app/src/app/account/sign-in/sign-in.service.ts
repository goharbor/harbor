import { Injectable } from '@angular/core';
import { Headers, Http, URLSearchParams } from '@angular/http';
import 'rxjs/add/operator/toPromise';

import { SignInCredential } from './sign-in-credential';

const url_prefix = '/ng';
const signInUrl = url_prefix + '/login';
/**
 * 
 * Define a service to provide sign in methods
 * 
 * @export
 * @class SignInService
 */
@Injectable()
export class SignInService {
    private headers = new Headers({
        "Content-Type": 'application/x-www-form-urlencoded'
    });

    constructor(private http: Http) {}

    //Handle the related exceptions
    private handleError(error: any): Promise<any>{
        return Promise.reject(error.message || error);
    }

    //Submit signin form to backend (NOT restful service)
    signIn(signInCredential: SignInCredential): Promise<any>{
        //Build the form package
        const body = new URLSearchParams();
        body.set('principal', signInCredential.principal);
        body.set('password', signInCredential.password);

        //Trigger Http
        return this.http.post(signInUrl, body.toString(), { headers: this.headers })
        .toPromise()
        .then(()=>null)
        .catch(this.handleError);
    }
}