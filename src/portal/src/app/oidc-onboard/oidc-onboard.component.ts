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
import { Router, ActivatedRoute } from '@angular/router';
import { Component, OnInit } from '@angular/core';
import { OidcOnboardService } from './oidc-onboard.service';
import { UntypedFormControl } from '@angular/forms';
import { CommonRoutes } from '../shared/entities/shared.const';
import { errorHandler } from '../shared/units/shared.utils';

@Component({
    selector: 'app-oidc-onboard',
    templateUrl: './oidc-onboard.component.html',
    styleUrls: ['./oidc-onboard.component.scss'],
})
export class OidcOnboardComponent implements OnInit {
    url: string;
    redirectUrl: string;
    errorMessage: string = '';
    oidcUsername = new UntypedFormControl('');
    errorOpen: boolean = false;
    constructor(
        private oidcOnboardService: OidcOnboardService,
        private router: Router,
        private route: ActivatedRoute
    ) {}

    ngOnInit() {
        this.route.queryParams.subscribe(params => {
            this.redirectUrl = params['redirect_url'] || '';
            this.oidcUsername.setValue(params['username'] || '');
        });
    }
    clickSaveBtn(): void {
        this.oidcOnboardService
            .oidcSave({ username: this.oidcUsername.value })
            .subscribe(
                res => {
                    if (this.redirectUrl === '') {
                        // Routing to the default location
                        this.router.navigateByUrl(CommonRoutes.HARBOR_DEFAULT);
                    } else {
                        this.router.navigateByUrl(this.redirectUrl);
                    }
                },
                error => {
                    this.errorMessage = errorHandler(error);
                    this.errorOpen = true;
                }
            );
    }
    emptyErrorMessage() {
        this.errorOpen = false;
    }
    backHarborPage() {
        this.router.navigate([CommonRoutes.HARBOR_DEFAULT]);
    }
}
