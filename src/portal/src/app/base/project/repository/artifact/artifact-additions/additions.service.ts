import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import {
    DEFAULT_SBOM_SUPPORTED_MIME_TYPES,
    DEFAULT_SUPPORTED_MIME_TYPES,
} from '../../../../../shared/units/utils';
import { ScanTypes } from 'src/app/shared/entities/shared.const';

@Injectable({
    providedIn: 'root',
})
export class AdditionsService {
    constructor(private http: HttpClient) {}

    getDetailByLink(
        link: string,
        shouldSetHeader: boolean,
        shouldReturnText: boolean,
        scanType = ScanTypes.VULNERABILITY
    ): Observable<any> {
        if (shouldReturnText) {
            return this.http.get(link, {
                observe: 'body',
                responseType: 'text',
            });
        }
        if (shouldSetHeader) {
            return this.http.get(link, {
                headers:
                    scanType === ScanTypes.SBOM
                        ? {
                              'X-Accept-SBOMs':
                                  DEFAULT_SBOM_SUPPORTED_MIME_TYPES,
                          }
                        : {
                              'X-Accept-Vulnerabilities':
                                  DEFAULT_SUPPORTED_MIME_TYPES,
                          },
            });
        }
        return this.http.get(link);
    }
}
