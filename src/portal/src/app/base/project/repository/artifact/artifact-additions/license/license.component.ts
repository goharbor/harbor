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
import { Component, Input, OnInit } from '@angular/core';
import { AdditionsService } from '../additions.service';
import { AdditionLink } from '../../../../../../../../ng-swagger-gen/models/addition-link';
import { ErrorHandler } from '../../../../../../shared/units/error-handler';
import { finalize } from 'rxjs/operators';

@Component({
    selector: 'hbr-artifact-license',
    templateUrl: './license.component.html',
    styleUrls: ['./license.component.scss'],
})
export class ArtifactLicenseComponent implements OnInit {
    @Input() licenseLink: AdditionLink;
    license: string;
    loading: boolean = false;
    fileTooLargeStatus: boolean = false;
    noLicenseStatus: boolean = false;

    constructor(
        private errorHandler: ErrorHandler,
        private additionsService: AdditionsService
    ) {}

    ngOnInit(): void {
        this.getLicense();
    }

    getLicense() {
        if (
            this.licenseLink &&
            !this.licenseLink.absolute &&
            this.licenseLink.href
        ) {
            this.loading = true;
            this.additionsService
                .getDetailByLink(this.licenseLink.href, false, true)
                .pipe(finalize(() => (this.loading = false)))
                .subscribe(
                    res => {
                        this.license = res;
                    },
                    error => {
                        if (error.status === 404) {
                            this.noLicenseStatus = true;
                        } else if (error.status === 413) {
                            this.fileTooLargeStatus = true;
                        } else {
                            this.errorHandler.error(error);
                        }
                    }
                );
        }
    }
}
