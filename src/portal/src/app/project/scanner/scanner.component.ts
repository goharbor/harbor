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
import { Component, OnInit, ViewChild } from "@angular/core";
import { ConfigScannerService } from "../../config/scanner/config-scanner.service";
import { Scanner } from "../../config/scanner/scanner";
import { MessageHandlerService } from "../../shared/message-handler/message-handler.service";
import { ErrorHandler } from "@harbor/ui";
import { ActivatedRoute } from "@angular/router";
import { ClrLoadingState } from "@clr/angular";
import { InlineAlertComponent } from "../../shared/inline-alert/inline-alert.component";
import { finalize } from "rxjs/operators";


@Component({
    selector: 'scanner',
    templateUrl: './scanner.component.html',
    styleUrls: ['./scanner.component.scss']
})
export class ScannerComponent implements OnInit {
    loading: boolean = false;
    scanners: Scanner[];
    scanner: Scanner;
    projectId: number;
    opened: boolean = false;
    selectedScanner: Scanner;
    saveBtnState: ClrLoadingState = ClrLoadingState.DEFAULT;
    onSaving: boolean = false;
    @ViewChild(InlineAlertComponent, { static: false }) inlineAlert: InlineAlertComponent;
    constructor( private configScannerService: ConfigScannerService,
                 private msgHandler: MessageHandlerService,
                 private errorHandler: ErrorHandler,
                 private route: ActivatedRoute,
    ) {
    }
    ngOnInit() {
        this.projectId = +this.route.snapshot.parent.params['id'];
        this.init();
    }
    init() {
        this.getScanner();
        this.getScanners();
    }
    getScanner() {
        this.configScannerService.getProjectScanner(this.projectId)
            .subscribe(response => {
                if (response && "{}" !== JSON.stringify(response)) {
                    this.scanner = response;
                    this.getScannerMetadata();
                }
            }, error => {
                this.errorHandler.error(error);
            });
    }
    getScannerMetadata() {
        if (this.scanner && this.scanner.uuid) {
            this.scanner.loadingMetadata = true;
            this.configScannerService.getScannerMetadata(this.scanner.uuid)
                .pipe(finalize(() => this.scanner.loadingMetadata = false))
                .subscribe(response => {
                    this.scanner.metadata = response;
                }, error => {
                    this.scanner.metadata = null;
                });
        }
    }
    getScanners() {
        this.loading = true;
        this.configScannerService.getScanners()
            .pipe(finalize(() => this.loading = false))
            .subscribe(response => {
                if (response && response.length > 0) {
                    this.scanners = response.filter(scanner => {
                       return !scanner.disabled;
                   });
                }
            }, error => {
                this.errorHandler.error(error);
            });
    }
    getMetadataForAll() {
        if (this.scanners && this.scanners.length > 0) {
            this.scanners.forEach((scanner, index) => {
                if (scanner.uuid ) {
                    this.scanners[index].loadingMetadata = true;
                    this.configScannerService.getScannerMetadata(scanner.uuid)
                        .pipe(finalize(() => this.scanners[index].loadingMetadata = false))
                        .subscribe(response => {
                            this.scanners[index].metadata = response;
                        }, error => {
                            this.scanners[index].metadata = null;
                        });
                }
            });
        }
    }
    close() {
        this.opened = false;
        this.selectedScanner = null;
    }
    open() {
        this.opened = true;
        this.inlineAlert.close();
        this.scanners.forEach(s => {
            if (this.scanner && s.uuid === this.scanner.uuid) {
                this.selectedScanner = s;
            }
        });
        this.getMetadataForAll();
    }
    get valid(): boolean {
        return this.selectedScanner
            && !(this.scanner && this.scanner.uuid === this.selectedScanner.uuid);
    }
    save() {
        this.saveBtnState = ClrLoadingState.LOADING;
        this.configScannerService.updateProjectScanner(this.projectId, this.selectedScanner.uuid)
            .subscribe(response => {
                this.close();
                this.msgHandler.showSuccess('Update Success');
                this.getScanner();
                this.saveBtnState = ClrLoadingState.SUCCESS;
            }, error => {
                this.inlineAlert.showInlineError(error);
                this.saveBtnState = ClrLoadingState.ERROR;
            });
    }
}
