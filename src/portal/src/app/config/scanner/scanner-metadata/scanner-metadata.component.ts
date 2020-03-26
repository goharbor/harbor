import {
    Component, Input,
    OnInit
} from "@angular/core";
import { ConfigScannerService } from "../config-scanner.service";
import { finalize } from "rxjs/operators";
import { ScannerMetadata } from "../scanner-metadata";
import { DatePipe } from "@angular/common";
import { TranslateService } from "@ngx-translate/core";
import { ErrorHandler } from "../../../../lib/utils/error-handler";
import { DATABASE_UPDATED_PROPERTY } from "../../../../lib/utils/utils";

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
                private translate: TranslateService) {
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
    parseDate(item: any): string {
        if (item && item.value && item.key === DATABASE_UPDATED_PROPERTY) {
            return new DatePipe(this.translate.currentLang).transform(item.value, 'short');
        }
        if (item && item.value) {
            return item.value;
        }
        return '';
    }
    toString(arr: string[]) {
        if (arr && arr.length > 0) {
            return "[" + arr.join(" , ") + "]";
        }
        return arr;
    }
}
