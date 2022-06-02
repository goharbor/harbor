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
import { Component, OnInit, ViewChild } from '@angular/core';
import { Scanner } from '../../left-side-nav/interrogation-services/scanner/scanner';
import { MessageHandlerService } from '../../../shared/services/message-handler.service';
import { ActivatedRoute } from '@angular/router';
import { ClrLoadingState } from '@clr/angular';
import { finalize } from 'rxjs/operators';
import { TranslateService } from '@ngx-translate/core';
import { ErrorHandler } from '../../../shared/units/error-handler';
import {
    UserPermissionService,
    USERSTATICPERMISSION,
} from '../../../shared/services';
import { InlineAlertComponent } from '../../../shared/components/inline-alert/inline-alert.component';
import { ProjectService } from '../../../../../ng-swagger-gen/services/project.service';
import { DEFAULT_PAGE_SIZE } from '../../../shared/units/utils';
import { forkJoin, Observable } from 'rxjs';
import { Project } from '../../../../../ng-swagger-gen/models/project';

@Component({
    selector: 'scanner',
    templateUrl: './scanner.component.html',
    styleUrls: ['./scanner.component.scss'],
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
    @ViewChild(InlineAlertComponent) inlineAlert: InlineAlertComponent;
    constructor(
        private msgHandler: MessageHandlerService,
        private errorHandler: ErrorHandler,
        private route: ActivatedRoute,
        private userPermissionService: UserPermissionService,
        private translate: TranslateService,
        private projectService: ProjectService
    ) {}
    ngOnInit() {
        this.projectId = +this.route.snapshot.parent.parent.params['id'];
        this.getPermission();
        this.init();
    }
    getPermission() {
        if (this.projectId) {
            this.userPermissionService
                .getPermission(
                    this.projectId,
                    USERSTATICPERMISSION.SCANNER.KEY,
                    USERSTATICPERMISSION.SCANNER.VALUE.CREATE
                )
                .subscribe(permission => {
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
        this.projectService
            .getScannerOfProject({
                projectNameOrId: this.projectId.toString(),
            })
            .pipe(finalize(() => (this.loading = false)))
            .subscribe(
                response => {
                    if (response && '{}' !== JSON.stringify(response)) {
                        this.scanner = response;
                        if (
                            isCheckHealth &&
                            this.scanner.health !== 'healthy'
                        ) {
                            this.translate
                                .get('SCANNER.SET_UNHEALTHY_SCANNER', {
                                    name: this.scanner.name,
                                })
                                .subscribe(res => {
                                    this.errorHandler.warning(res);
                                });
                        }
                    }
                },
                error => {
                    this.errorHandler.error(error);
                }
            );
    }
    getScanners() {
        if (this.projectId) {
            this.projectService
                .listScannerCandidatesOfProjectResponse({
                    projectNameOrId: this.projectId.toString(),
                    page: 1,
                    pageSize: DEFAULT_PAGE_SIZE,
                })
                .subscribe(response => {
                    if (response.headers) {
                        const xHeader: string =
                            response.headers.get('X-Total-Count');
                        const totalCount = parseInt(xHeader, 0);
                        let arr = response.body || [];
                        if (totalCount <= DEFAULT_PAGE_SIZE) {
                            // already gotten all scanners
                            if (arr && arr.length > 0) {
                                this.scanners = arr.filter(scanner => {
                                    return !scanner.disabled;
                                });
                            }
                        } else {
                            // get all the scanners in specified times
                            const times: number = Math.ceil(
                                totalCount / DEFAULT_PAGE_SIZE
                            );
                            const observableList: Observable<Project[]>[] = [];
                            for (let i = 2; i <= times; i++) {
                                observableList.push(
                                    this.projectService.listScannerCandidatesOfProject(
                                        {
                                            page: i,
                                            pageSize: DEFAULT_PAGE_SIZE,
                                            projectNameOrId:
                                                this.projectId.toString(),
                                        }
                                    )
                                );
                            }
                            forkJoin(observableList).subscribe(res => {
                                if (res && res.length) {
                                    res.forEach(item => {
                                        arr = arr.concat(item);
                                    });
                                    this.scanners = arr.filter(scanner => {
                                        return !scanner.disabled;
                                    });
                                }
                            });
                        }
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
        return (
            this.selectedScanner &&
            !(this.scanner && this.scanner.uuid === this.selectedScanner.uuid)
        );
    }
    save() {
        this.saveBtnState = ClrLoadingState.LOADING;
        this.projectService
            .setScannerOfProject({
                projectNameOrId: this.projectId.toString(),
                payload: {
                    uuid: this.selectedScanner.uuid,
                },
            })
            .subscribe(
                response => {
                    this.close();
                    this.msgHandler.showSuccess('Update Success');
                    this.getScanner(true);
                    this.saveBtnState = ClrLoadingState.SUCCESS;
                },
                error => {
                    this.inlineAlert.showInlineError(error);
                    this.saveBtnState = ClrLoadingState.ERROR;
                }
            );
    }
}
