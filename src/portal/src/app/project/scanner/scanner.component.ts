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
import { ActivatedRoute } from "@angular/router";
import { ClrLoadingState } from "@clr/angular";
import { InlineAlertComponent } from "../../shared/inline-alert/inline-alert.component";
import { finalize } from "rxjs/operators";
import { TranslateService } from "@ngx-translate/core";
import { ErrorHandler } from "../../../lib/utils/error-handler";
import { UserPermissionService, USERSTATICPERMISSION } from "../../../lib/services";


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
    hasCreatePermission: boolean = false;
    @ViewChild(InlineAlertComponent, { static: false }) inlineAlert: InlineAlertComponent;
    constructor( private configScannerService: ConfigScannerService,
                 private msgHandler: MessageHandlerService,
                 private errorHandler: ErrorHandler,
                 private route: ActivatedRoute,
                 private userPermissionService: UserPermissionService,
                 private translate: TranslateService
    ) {
    }
    ngOnInit() {
        this.projectId = +this.route.snapshot.parent.params['id'];
        this.getPermission();
        this.init();
    }
    getPermission() {
        if (this.projectId) {
            this.userPermissionService.getPermission(this.projectId,
                USERSTATICPERMISSION.SCANNER.KEY, USERSTATICPERMISSION.SCANNER.VALUE.CREATE)
                .subscribe( permission => {
                    this.hasCreatePermission = permission;
                    if (this.hasCreatePermission) {
                        this.getScanners();
                    }
                 });
        }
    }
    init() {
        this.getScanner();
    }
    getScanner(isCheckHealth?: boolean) {
        this.loading = true;
        this.configScannerService.getProjectScanner(this.projectId)
            .pipe(finalize(() => this.loading = false))
            .subscribe(response => {
                if (response && "{}" !== JSON.stringify(response)) {
                    this.scanner = response;
                    if (isCheckHealth && this.scanner.health !== 'healthy') {
                        this.translate.get("SCANNER.SET_UNHEALTHY_SCANNER", {name: this.scanner.name})
                            .subscribe(res => {
                                 this.errorHandler.warning(res);
                            }
                        );
                    }
                }
            }, error => {
                this.errorHandler.error(error);
            });
    }
    getScanners() {
        if (this.projectId) {
            this.configScannerService.getProjectScanners(this.projectId)
                .subscribe(response => {
                    if (response && response.length > 0) {
                        this.scanners = response.filter(scanner => {
                            return !scanner.disabled;
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
                this.getScanner(true);
                this.saveBtnState = ClrLoadingState.SUCCESS;
            }, error => {
                this.inlineAlert.showInlineError(error);
                this.saveBtnState = ClrLoadingState.ERROR;
            });
    }
}
