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
import { DatePipe } from '@angular/common';
import { ErrorHandler } from '../../../../../shared/units/error-handler';
import {
    DATABASE_NEXT_UPDATE_PROPERTY,
    DATABASE_UPDATED_PROPERTY,
} from '../../../../../shared/units/utils';
import { OptimizerService } from '../../../../../../../ng-swagger-gen/services/optimizer.service';
import { OptimizerAdapterMetadata } from '../../../../../../../ng-swagger-gen/models/optimizer-adapter-metadata';

@Component({
    selector: 'optimizer-metadata',
    templateUrl: 'optimizer-metadata.html',
    styleUrls: ['./optimizer-metadata.scss'],
    standalone: false,
})
export class OptimizerMetadataComponent implements OnInit {
    @Input() uid: string;
    loading: boolean = false;
    optimizerMetadata: OptimizerAdapterMetadata;
    constructor(
        private configOptimizerService: OptimizerService,
        private errorHandler: ErrorHandler
    ) {}
    ngOnInit(): void {
        this.loading = true;
        this.configOptimizerService
            .getOptimizerMetadata({
                registrationId: this.uid,
            })
            .pipe(finalize(() => (this.loading = false)))
            .subscribe(
                response => {
                    this.optimizerMetadata = response;
                },
                error => {
                    this.errorHandler.error(error);
                }
            );
    }
    parseDate(item: any): string {
        if (this.hasValue(item) && this.hasDateValue(item)) {
            return new DatePipe('en-us').transform(item.value, 'short');
        }
        if (this.hasValue(item)) {
            return item.value;
        }
        return '';
    }
    hasValue(item: any): boolean {
        return item && item.value;
    }
    hasDateValue(item: any): boolean {
        switch (item.key) {
            case DATABASE_UPDATED_PROPERTY:
            case DATABASE_NEXT_UPDATE_PROPERTY:
                return true;
            default:
                return false;
        }
    }
    toString(arr: string[]) {
        if (arr && arr.length > 0) {
            return '[' + arr.join(' , ') + ']';
        }
        return arr;
    }
}
