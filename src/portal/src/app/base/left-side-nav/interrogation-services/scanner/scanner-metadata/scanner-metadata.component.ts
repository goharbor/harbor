import { Component, Input, OnInit } from '@angular/core';
import { finalize } from 'rxjs/operators';
import { DatePipe } from '@angular/common';
import { ErrorHandler } from '../../../../../shared/units/error-handler';
import {
    DATABASE_NEXT_UPDATE_PROPERTY,
    DATABASE_UPDATED_PROPERTY,
} from '../../../../../shared/units/utils';
import { ScannerService } from '../../../../../../../ng-swagger-gen/services/scanner.service';
import { ScannerAdapterMetadata } from '../../../../../../../ng-swagger-gen/models/scanner-adapter-metadata';

@Component({
    selector: 'scanner-metadata',
    templateUrl: 'scanner-metadata.html',
    styleUrls: ['./scanner-metadata.scss'],
})
export class ScannerMetadataComponent implements OnInit {
    @Input() uid: string;
    loading: boolean = false;
    scannerMetadata: ScannerAdapterMetadata;
    constructor(
        private configScannerService: ScannerService,
        private errorHandler: ErrorHandler
    ) {}
    ngOnInit(): void {
        this.loading = true;
        this.configScannerService
            .getScannerMetadata({
                registrationId: this.uid,
            })
            .pipe(finalize(() => (this.loading = false)))
            .subscribe(
                response => {
                    this.scannerMetadata = response;
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
