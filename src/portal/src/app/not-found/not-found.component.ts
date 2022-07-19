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
import { Component, OnInit, OnDestroy } from '@angular/core';
import { Router } from '@angular/router';

const defaultInterval = 1000;
const defaultLeftTime = 5;

@Component({
    selector: 'page-not-found',
    templateUrl: 'not-found.component.html',
    styleUrls: ['not-found.component.scss'],
})
export class PageNotFoundComponent implements OnInit, OnDestroy {
    leftSeconds: number = defaultLeftTime;
    timeInterval: any = null;

    constructor(private router: Router) {}

    ngOnInit(): void {
        if (!this.timeInterval) {
            this.timeInterval = setInterval(interval => {
                this.leftSeconds--;
                if (this.leftSeconds <= 0) {
                    this.router.navigate(['harbor']);
                    clearInterval(this.timeInterval);
                }
            }, defaultInterval);
        }
    }

    ngOnDestroy(): void {
        if (this.timeInterval) {
            clearInterval(this.timeInterval);
        }
    }
}
