import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { DEFAULT_SUPPORTED_MIME_TYPES } from '../../../../../shared/units/utils';

@Injectable({
    providedIn: 'root',
})
export class AdditionsService {
    constructor(private http: HttpClient) {}

    getDetailByLink(
        link: string,
        shouldSetHeader: boolean,
        shouldReturnText: boolean
    ): Observable<any> {
        if (shouldReturnText) {
            return this.http.get(link, {
                observe: 'body',
                responseType: 'text',
            });
        }
        if (shouldSetHeader) {
            return this.http.get(link, {
                headers: {
                    'X-Accept-Vulnerabilities': DEFAULT_SUPPORTED_MIME_TYPES,
                },
            });
        }
        return this.http.get(link);
    }
}
