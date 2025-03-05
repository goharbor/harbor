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
import { ArtifactDependency } from '../models';
import { AdditionsService } from '../additions.service';
import { AdditionLink } from '../../../../../../../../ng-swagger-gen/models/addition-link';
import { ErrorHandler } from '../../../../../../shared/units/error-handler';
import { finalize } from 'rxjs/operators';

@Component({
    selector: 'hbr-artifact-dependencies',
    templateUrl: './dependencies.component.html',
    styleUrls: ['./dependencies.component.scss'],
})
export class DependenciesComponent implements OnInit {
    @Input()
    dependenciesLink: AdditionLink;
    dependencyList: ArtifactDependency[] = [];
    loading: boolean = false;
    constructor(
        private errorHandler: ErrorHandler,
        private additionsService: AdditionsService
    ) {}

    ngOnInit(): void {
        this.getDependencyList();
    }
    getDependencyList() {
        if (
            this.dependenciesLink &&
            !this.dependenciesLink.absolute &&
            this.dependenciesLink.href
        ) {
            this.loading = true;
            this.additionsService
                .getDetailByLink(this.dependenciesLink.href, false, false)
                .pipe(finalize(() => (this.loading = false)))
                .subscribe(
                    res => {
                        this.dependencyList = res;
                    },
                    error => {
                        this.errorHandler.error(error);
                    }
                );
        }
    }
}
