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
import { finalize } from 'rxjs/operators';
import { AdditionLink } from '../../../../../../../../ng-swagger-gen/models/addition-link';
import { ErrorHandler } from '../../../../../../shared/units/error-handler';
import { downloadJson } from '../../../../../../shared/units/utils';
import { AdditionsService } from '../additions.service';

@Component({
    selector: 'hbr-artifact-vex',
    templateUrl: './artifact-vex.component.html',
    styleUrls: ['./artifact-vex.component.scss'],
    standalone: false,
})
export class ArtifactVEXComponent implements OnInit {
    @Input()
    vexLink: AdditionLink;

    content: string;
    loading: boolean = false;
    error: boolean = false;
    private contentObj: object | null;

    constructor(
        private errorHandler: ErrorHandler,
        private additionsService: AdditionsService
    ) {}

    ngOnInit(): void {
        if (
            this.vexLink &&
            !this.vexLink.absolute &&
            this.vexLink.href
        ) {
            this.loading = true;
            this.additionsService
                .getDetailByLink(this.vexLink.href, false, true)
                .pipe(finalize(() => (this.loading = false)))
                .subscribe(
                    res => {
                        if (res) {
                            if (typeof res === 'object') {
                                this.contentObj = res;
                                this.content = JSON.stringify(res, null, 2);
                            } else {
                                try {
                                    this.contentObj = JSON.parse(res);
                                    this.content = JSON.stringify(
                                        this.contentObj,
                                        null,
                                        2
                                    );
                                } catch {
                                    this.contentObj = null;
                                    this.content = res;
                                }
                            }
                        }
                    },
                    error => {
                        this.error = true;
                        this.errorHandler.error(error);
                    }
                );
        }
    }

    download(): void {
        if (this.contentObj) {
            downloadJson(this.contentObj, 'vex.json');
        }
    }
}