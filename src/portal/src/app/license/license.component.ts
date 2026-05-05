// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
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
