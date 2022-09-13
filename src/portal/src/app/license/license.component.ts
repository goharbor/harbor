import { Component, OnInit } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { throwError as observableThrowError } from 'rxjs';
import { catchError } from 'rxjs/operators';
import { Title } from '@angular/platform-browser';
@Component({
    selector: 'app-license',
    viewProviders: [Title],
    templateUrl: './license.component.html',
    styleUrls: ['./license.component.scss'],
})
export class LicenseComponent implements OnInit {
    constructor(private http: HttpClient) {}
    public licenseContent: any;
    ngOnInit() {
        this.http
            .get('/LICENSE', { responseType: 'text' })
            .pipe(catchError(error => observableThrowError(error)))
            .subscribe(json => {
                this.licenseContent = json;
            });
    }
}
