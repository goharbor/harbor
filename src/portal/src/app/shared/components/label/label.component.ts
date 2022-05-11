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
import { Component, ViewChild, Input } from '@angular/core';
import { ErrorHandler } from '../../units/error-handler';
import { CreateEditLabelComponent } from './create-edit-label/create-edit-label.component';
import {
    ConfirmationButtons,
    ConfirmationState,
    ConfirmationTargets,
} from '../../entities/shared.const';
import { TranslateService } from '@ngx-translate/core';
import { ConfirmationDialogComponent } from '../confirmation-dialog';
import {
    operateChanges,
    OperateInfo,
    OperationState,
} from '../operation/operate';
import { OperationService } from '../operation/operation.service';
import { map, catchError, finalize } from 'rxjs/operators';
import { Observable, throwError as observableThrowError, forkJoin } from 'rxjs';
import { errorHandler } from '../../units/shared.utils';
import { ConfirmationMessage } from '../../../base/global-confirmation-dialog/confirmation-message';
import { ConfirmationAcknowledgement } from '../../../base/global-confirmation-dialog/confirmation-state-message';
import { LabelService } from '../../../../../ng-swagger-gen/services/label.service';
import { Label } from '../../../../../ng-swagger-gen/models/label';
import {
    getPageSizeFromLocalStorage,
    getSortingString,
    PageSizeMapKeys,
    setPageSizeToLocalStorage,
} from '../../units/utils';
import { ClrDatagridStateInterface } from '@clr/angular';

@Component({
    selector: 'hbr-label',
    templateUrl: './label.component.html',
    styleUrls: ['./label.component.scss'],
})
export class LabelComponent {
    timerHandler: any;
    loading: boolean = true;
    targets: Label[];
    targetName: string;
    selectedRow: Label[] = [];

    @Input() scope: string;
    @Input() projectId = 0;
    @Input() hasCreateLabelPermission: boolean;
    @Input() hasUpdateLabelPermission: boolean;
    @Input() hasDeleteLabelPermission: boolean;

    @ViewChild(CreateEditLabelComponent)
    createEditLabel: CreateEditLabelComponent;
    @ViewChild('confirmationDialog')
    confirmationDialogComponent: ConfirmationDialogComponent;

    page: number = 1;
    pageSize: number = getPageSizeFromLocalStorage(
        PageSizeMapKeys.LABEL_COMPONENT
    );
    total: number = 0;
    constructor(
        private labelService: LabelService,
        private errorHandlerEntity: ErrorHandler,
        private translateService: TranslateService,
        private operationService: OperationService
    ) {}

    retrieve(state?: ClrDatagridStateInterface) {
        this.selectedRow = [];
        // this.targetName = "";
        if (state && state.page) {
            this.pageSize = state.page.size;
            setPageSizeToLocalStorage(
                PageSizeMapKeys.LABEL_COMPONENT,
                this.pageSize
            );
        }
        let sort: string;
        if (state && state.sort && state.sort.by) {
            sort = getSortingString(state);
        } else {
            // sort by creation_time desc by default
            sort = `-creation_time`;
        }
        this.loading = true;
        this.labelService
            .ListLabelsResponse({
                page: this.page,
                pageSize: this.pageSize,
                name: this.targetName,
                sort: sort,
                scope: this.scope,
                projectId: this.projectId,
            })
            .pipe(
                finalize(() => {
                    this.loading = false;
                })
            )
            .subscribe(
                res => {
                    // Get total count
                    if (res.headers) {
                        let xHeader: string = res.headers.get('X-Total-Count');
                        if (xHeader) {
                            this.total = parseInt(xHeader, 0);
                        }
                    }
                    this.targets = res.body || [];
                },
                error => {
                    this.errorHandlerEntity.error(error);
                }
            );
    }

    openModal(): void {
        this.createEditLabel.openModal();
    }

    reload(): void {
        this.targetName = '';
        this.page = 1;
        this.total = 0;
        this.selectedRow = [];
        this.retrieve();
    }

    doSearchTargets(targetName: string) {
        this.targetName = targetName;
        this.page = 1;
        this.total = 0;
        this.selectedRow = [];
        this.retrieve();
    }

    refreshTargets() {
        this.targetName = '';
        this.page = 1;
        this.total = 0;
        this.selectedRow = [];
        this.retrieve();
    }
    editLabel(label: Label[]): void {
        this.createEditLabel.editModel(label[0].id, label);
    }

    deleteLabels(targets: Label[]): void {
        if (targets && targets.length) {
            let targetNames: string[] = [];
            targets.forEach(target => {
                targetNames.push(target.name);
            });
            let deletionMessage = new ConfirmationMessage(
                'LABEL.DELETION_TITLE_TARGET',
                'LABEL.DELETION_SUMMARY_TARGET',
                targetNames.join(', ') || '',
                targets,
                ConfirmationTargets.TARGET,
                ConfirmationButtons.DELETE_CANCEL
            );
            this.confirmationDialogComponent.open(deletionMessage);
        }
    }

    confirmDeletion(message: ConfirmationAcknowledgement) {
        if (
            message &&
            message.source === ConfirmationTargets.TARGET &&
            message.state === ConfirmationState.CONFIRMED
        ) {
            let targetLists: Label[] = message.data;
            if (targetLists && targetLists.length) {
                let observableLists: any[] = [];
                targetLists.forEach(target => {
                    observableLists.push(this.delOperate(target));
                });
                forkJoin(...observableLists).subscribe(
                    item => {
                        this.reload();
                    },
                    error => {
                        this.errorHandlerEntity.error(error);
                    }
                );
            }
        }
    }

    delOperate(target: Label): Observable<any> {
        // init operation info
        let operMessage = new OperateInfo();
        operMessage.name = 'OPERATION.DELETE_LABEL';
        operMessage.data.id = target.id;
        operMessage.state = OperationState.progressing;
        operMessage.data.name = target.name;
        this.operationService.publishInfo(operMessage);

        return this.labelService.DeleteLabel({ labelId: target.id }).pipe(
            map(response => {
                this.translateService
                    .get('BATCH.DELETED_SUCCESS')
                    .subscribe(res => {
                        operateChanges(operMessage, OperationState.success);
                    });
            }),
            catchError(error => {
                const message = errorHandler(error);
                this.translateService
                    .get(message)
                    .subscribe(res =>
                        operateChanges(operMessage, OperationState.failure, res)
                    );
                return observableThrowError(error);
            })
        );
    }
}
