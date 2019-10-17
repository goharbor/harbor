import {
    Component, Inject, Input, LOCALE_ID,
    OnInit
} from "@angular/core";
import { ConfigScannerService } from "../config-scanner.service";
import { finalize } from "rxjs/operators";
import { ErrorHandler } from "@harbor/ui";
import { ScannerMetadata } from "../scanner-metadata";
import { DatePipe } from "@angular/common";

@Component({
    selector: 'scanner-metadata',
    templateUrl: 'scanner-metadata.html',
    styleUrls: ['./scanner-metadata.scss']
})
export class ScannerMetadataComponent implements  OnInit {
    @Input() uid: string;
    loading: boolean = false;
    scannerMetadata: ScannerMetadata;
    constructor(private configScannerService: ConfigScannerService,
                private errorHandler: ErrorHandler,
    @Inject(LOCALE_ID) private _locale: string) {
    }
    ngOnInit(): void {
        this.loading = true;
        this.configScannerService.getScannerMetadata(this.uid)
            .pipe(finalize(() => this.loading = false))
            .subscribe(response => {
                this.scannerMetadata = response;
            }, error => {
                this.errorHandler.error(error);
            });
    }
    parseDate(str: string): string {
        try {
            if (str === new Date(str).toISOString()) {
                return new DatePipe(this._locale).transform(str, 'short');
            }
        } catch (e) {
            return str;
        }
        return str;
    }
    toString(arr: string[]) {
        if (arr && arr.length > 0) {
            return "[" + arr.join(" , ") + "]";
        }
        return arr;
    }
}
