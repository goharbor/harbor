import { Injectable } from '@angular/core';
import {
    HttpInterceptor,
    HttpRequest,
    HttpHandler,
    HttpResponse,
    HttpErrorResponse,
} from '@angular/common/http';
import { Observable, throwError } from 'rxjs';
import { catchError, tap } from 'rxjs/operators';

export const SAFE_METHODS: string[] = ['GET', 'HEAD', 'OPTIONS', 'TRACE'];

@Injectable({
    providedIn: 'root',
})
export class InterceptHttpService implements HttpInterceptor {
    constructor() {}

    intercept(request: HttpRequest<any>, next: HttpHandler): Observable<any> {
        // Get the csrf token from localstorage
        const token = localStorage.getItem('__csrf');
        if (token) {
            // Clone the request and replace the original headers with
            // cloned headers, updated with the csrf token.
            // not for requests using safe methods
            if (
                request.method &&
                SAFE_METHODS.indexOf(request.method.toUpperCase()) === -1
            ) {
                request = request.clone({
                    headers: request.headers.set('X-Harbor-CSRF-Token', token),
                });
            }
        }
        return next
            .handle(request)
            .pipe(
                tap(
                    response => {
                        if (
                            response &&
                            response instanceof HttpResponse &&
                            response.headers
                        ) {
                            const responseToken: string = response.headers.get(
                                'X-Harbor-CSRF-Token'
                            );
                            if (responseToken) {
                                localStorage.setItem('__csrf', responseToken);
                            }
                        }
                    },
                    error => {
                        if (error && error.headers) {
                            const responseToken: string = error.headers.get(
                                'X-Harbor-CSRF-Token'
                            );
                            if (responseToken) {
                                localStorage.setItem('__csrf', responseToken);
                            }
                        }
                    }
                )
            )
            .pipe(
                catchError(error => {
                    // handle 504 error in document format from backend
                    if (error && error.status === 504) {
                        // throw 504 error in json format
                        return throwError(
                            new HttpErrorResponse({
                                error: '504 gateway timeout',
                                status: 504,
                            })
                        );
                    }
                    if (error.status === 403) {
                        const csrfToken = localStorage.getItem('__csrf');
                        if (csrfToken) {
                            request = request.clone({
                                headers: request.headers.set(
                                    'X-Harbor-CSRF-Token',
                                    csrfToken
                                ),
                            });
                            return next.handle(request);
                        }
                    }
                    return throwError(error);
                })
            );
    }
}
