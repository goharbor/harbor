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
import { ArtifactBuildHistory } from '../models';
import { finalize } from 'rxjs/operators';
import { AdditionLink } from '../../../../../../../../ng-swagger-gen/models/addition-link';
import { ErrorHandler } from '../../../../../../shared/units/error-handler';

@Component({
    selector: 'hbr-artifact-build-history',
    templateUrl: './build-history.component.html',
    styleUrls: ['./build-history.component.scss'],
})
export class BuildHistoryComponent implements OnInit {
    @Input()
    buildHistoryLink: AdditionLink;
    historyList: ArtifactBuildHistory[] = [];
    loading: Boolean = false;
    constructor(
        private errorHandler: ErrorHandler,
        private additionsService: AdditionsService
    ) {}

    ngOnInit(): void {
        this.getBuildHistory();
    }
    getBuildHistory() {
        if (
            this.buildHistoryLink &&
            !this.buildHistoryLink.absolute &&
            this.buildHistoryLink.href
        ) {
            this.loading = true;
            this.additionsService
                .getDetailByLink(this.buildHistoryLink.href, false, false)
                .pipe(finalize(() => (this.loading = false)))
                .subscribe(
                    res => {
                        if (res && res.length) {
                            res.forEach((ele: any) => {
                                const history: ArtifactBuildHistory =
                                    new ArtifactBuildHistory();
                                history.created = ele.created;
                                if (ele.created_by !== undefined) {
                                    let createdBy = ele.created_by
                                        .replace('/bin/sh -c #(nop)', '')
                                        .trimLeft();
                                    if (!createdBy.startsWith('RUN ')) {
                                        createdBy = createdBy.replace(
                                            '/bin/sh -c',
                                            'RUN'
                                        );
                                    }
                                    history.created_by = createdBy.replace(
                                        /\s+# buildkit$/,
                                        ''
                                    );
                                } else {
                                    history.created_by = ele.comment;
                                }
                                this.historyList.push(history);
                            });
                        }
                    },
                    error => {
                        this.errorHandler.error(error);
                    }
                );
        }
    }
}
