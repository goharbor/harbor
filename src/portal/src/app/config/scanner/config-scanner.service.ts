import {Injectable} from "@angular/core";
import {Scanner} from "./scanner";
import { forkJoin, Observable, of, throwError as observableThrowError } from "rxjs";
import { catchError, delay, map } from "rxjs/operators";
import { HttpClient } from "@angular/common/http";
import { ScannerMetadata } from "./scanner-metadata";

@Injectable()
export class ConfigScannerService {

    constructor( private http: HttpClient) {}
    getScannersByName(name: string): Observable<Scanner[]> {
            return this.http.get(`/api/scanners?q=name=${name}`)
                .pipe(map(response => response as Scanner[]));
    }
    getScannersByEndpointUrl(endpointUrl: string): Observable<Scanner[]> {
        return this.http.get(`/api/scanners?q=url=${endpointUrl}`)
            .pipe(map(response => response as Scanner[]));
    }
    testEndpointUrl(endpointUrl: string): Observable<any> {
        if (endpointUrl === 'http://196.168.1.1') {
            return of([new Scanner()]).pipe(delay(1500));
        }
        return of([]).pipe(delay(1500));
    }
    addScanner(scanner: Scanner): Observable<any> {
        return this.http.post('/api/scanners', scanner )
                .pipe(catchError(error => observableThrowError(error)));
    }
    getScanners(): Observable<Scanner[]> {
        return this.http.get('/api/scanners')
            .pipe(map(response => response as Scanner[]))
            .pipe(catchError(error => observableThrowError(error)));
    }
    updateScanner(scanner: Scanner): Observable<any> {
        return this.http.put(`/api/scanners/${scanner.uid}`, scanner )
            .pipe(catchError(error => observableThrowError(error)));
    }
    deleteScanner(scanner: Scanner): Observable<any> {
        return this.http.delete(`/api/scanners/${scanner.uid}`)
            .pipe(catchError(error => observableThrowError(error)));
    }
    deleteScanners(scanners: Scanner[]): Observable<any> {
        let observableLists: any[] = [];
        if (scanners && scanners.length > 0) {
            scanners.forEach(scanner => {
                observableLists.push(this.deleteScanner(scanner));
            });
            return forkJoin(...observableLists);
        }
    }
    getProjectScanner(projectId: number): Observable<Scanner>  {
        return this.http.get(`/api/projects/${projectId}/scanner`)
            .pipe(map(response => response as Scanner))
            .pipe(catchError(error => observableThrowError(error)));
    }
    updateProjectScanner(projectId: number , uid: string): Observable<any>  {
        return this.http.put(`/api/projects/${projectId}/scanner` , {uuid: uid})
            .pipe(catchError(error => observableThrowError(error)));
    }
    getSannerMetadate(uid: string): Observable<ScannerMetadata> {
        /*return of({
            "scanner": {
                "name": "Microscanner",
                "vendor": "Aqua Security",
                "version": "3.0.5"
            },
            "capabilities": [
                {
                    "consumes_mime_types": [
                        "application/vnd.oci.image.manifest.v1+json",
                        "application/vnd.docker.distribution.manifest.v2+json"
                    ],
                    "produces_mime_types": [
                        "application/vnd.scanner.adapter.vuln.report.harbor+json; version=1.0"
                    ]
                }
            ],
            "properties": {
                "harbor.scanner-adapter/scanner-type": "os-package-vulnerability",
                "harbor.scanner-adapter/vulnerability-database-updated-at": "2019-08-13T08:16:33.345Z"
            }})
            .pipe(map(response => response as ScannerMetadata), delay(1500));*/
        return this.http.get(`/api/sacnners/${uid}/metadata`)
            .pipe(map(response => response as ScannerMetadata))
            .pipe(catchError(error => observableThrowError(error)));
    }
}
