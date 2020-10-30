// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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
import {
    Component,
    Input,
    Output,
    EventEmitter,
    ViewChild,
} from "@angular/core";
import { TranslateService } from "@ngx-translate/core";
import { map, catchError, finalize } from "rxjs/operators";
import { Observable, forkJoin, throwError as observableThrowError } from "rxjs";
import {HELM_HUB, ReplicationService} from "../../services";
import {
    ReplicationRule
} from "../../services";
import { ConfirmationDialogComponent } from "../confirmation-dialog";
import { ConfirmationMessage } from "../confirmation-dialog";
import { ConfirmationAcknowledgement } from "../confirmation-dialog";
import {
    ConfirmationState,
    ConfirmationTargets,
    ConfirmationButtons
} from "../../entities/shared.const";
import { ErrorHandler } from "../../utils/error-handler";
import { clone } from "../../utils/utils";
import { operateChanges, OperateInfo, OperationState } from "../operation/operate";
import { OperationService } from "../operation/operation.service";
import { errorHandler as errorHandFn} from "../../utils/shared/shared.utils";
import { ClrDatagridStateInterface } from '@clr/angular';
@Component({
    selector: "hbr-list-replication-rule",
    templateUrl: "./list-replication-rule.component.html",
    styleUrls: ["./list-replication-rule.component.scss"],
})
export class ListReplicationRuleComponent  {
    @Input() selectedId: number | string;
    @Input() withReplicationJob: boolean;
    @Input() hasCreateReplicationPermission: boolean;
    @Input() hasUpdateReplicationPermission: boolean;
    @Input() hasDeleteReplicationPermission: boolean;
    @Input() hasExecuteReplicationPermission: boolean;
    @Output() selectOne = new EventEmitter<ReplicationRule>();
    @Output() editOne = new EventEmitter<ReplicationRule>();
    @Output() toggleOne = new EventEmitter<ReplicationRule>();
    @Output() hideJobs = new EventEmitter<any>();
    @Output() redirect = new EventEmitter<ReplicationRule>();
    @Output() openNewRule = new EventEmitter<any>();
    @Output() replicateManual = new EventEmitter<ReplicationRule>();
    rules: ReplicationRule[] = [];
    selectedRow: ReplicationRule;
    @ViewChild("toggleConfirmDialog")
    toggleConfirmDialog: ConfirmationDialogComponent;
    @ViewChild("deletionConfirmDialog")
    deletionConfirmDialog: ConfirmationDialogComponent;
    page: number = 1;
    pageSize: number = 5;
    totalCount: number = 0;
    ruleName: string = "";
    loading: boolean = true;

    constructor(private replicationService: ReplicationService,
        private translateService: TranslateService,
        private errorHandler: ErrorHandler,
        private operationService: OperationService) {
    }

    trancatedDescription(desc: string): string {
        if (desc.length > 35) {
            return desc.substr(0, 35);
        } else {
            return desc;
        }
    }
    replicateRule(rule: ReplicationRule): void {
        this.replicateManual.emit(rule);
    }
    deletionConfirm(message: ConfirmationAcknowledgement) {
        if (
            message &&
            message.source === ConfirmationTargets.POLICY &&
            message.state === ConfirmationState.CONFIRMED
        ) {
            this.deleteOpe(message.data);
        }
        if ( message &&
            message.source === ConfirmationTargets.REPLICATION &&
            message.state === ConfirmationState.CONFIRMED) {
            const rule: ReplicationRule = clone(message.data);
            rule.enabled = !message.data.enabled;
            const opeMessage = new OperateInfo();
            opeMessage.name = rule.enabled ? 'REPLICATION.ENABLE_TITLE' : 'REPLICATION.DISABLE_TITLE';
            opeMessage.data.id = rule.id;
            opeMessage.state = OperationState.progressing;
            opeMessage.data.name = rule.name;
            this.operationService.publishInfo(opeMessage);
            this.replicationService.updateReplicationRule(rule.id, rule).subscribe(
                res => {
                    this.translateService.get(rule.enabled ? 'REPLICATION.ENABLE_SUCCESS' : 'REPLICATION.DISABLE_SUCCESS')
                        .subscribe(msg => {
                        operateChanges(opeMessage, OperationState.success);
                        this.errorHandler.info(msg);
                        this.refreshRule();
                    });
                }, error => {
                    const errMessage = errorHandFn(error);
                    this.translateService.get(rule.enabled ? 'REPLICATION.ENABLE_FAILED' : 'REPLICATION.DISABLE_FAILED')
                        .subscribe(msg => {
                        operateChanges(opeMessage, OperationState.failure, msg);
                        this.errorHandler.error(errMessage);
                    });
                }
            );
        }
    }

    selectRule(rule: ReplicationRule): void {
        if (rule) {
            this.selectedId = rule.id || "";
            this.selectOne.emit(rule);
        }
    }

    redirectTo(rule: ReplicationRule): void {
        this.redirect.emit(rule);
    }

    openModal(): void {
        this.openNewRule.emit();
    }

    editRule(rule: ReplicationRule) {
        this.editOne.emit(rule);
    }

    deleteRule(rule: ReplicationRule) {
        if (rule) {
            let deletionMessage = new ConfirmationMessage(
                "REPLICATION.DELETION_TITLE",
                "REPLICATION.DELETION_SUMMARY",
                rule.name,
                rule,
                ConfirmationTargets.POLICY,
                ConfirmationButtons.DELETE_CANCEL
            );
            this.deletionConfirmDialog.open(deletionMessage);
        }
    }

    deleteOpe(rule: ReplicationRule) {
        if (rule) {
            let observableLists: any[] = [];
            observableLists.push(this.delOperate(rule));

            forkJoin(...observableLists).subscribe(item => {
                this.selectedRow = null;
                this.refreshRule();
            }, error => {
                this.errorHandler.error(error);
            });
        }
    }

    delOperate(rule: ReplicationRule): Observable<any> {
        // init operation info
        let operMessage = new OperateInfo();
        operMessage.name = 'OPERATION.DELETE_REPLICATION';
        operMessage.data.id = +rule.id;
        operMessage.state = OperationState.progressing;
        operMessage.data.name = rule.name;
        this.operationService.publishInfo(operMessage);

        return this.replicationService
            .deleteReplicationRule(+rule.id)
            .pipe(map(() => {
                this.translateService.get('BATCH.DELETED_SUCCESS')
                    .subscribe(res => operateChanges(operMessage, OperationState.success));
            })
                , catchError(error => {
                    const message = errorHandFn(error);
                    this.translateService.get(message).subscribe(res =>
                        operateChanges(operMessage, OperationState.failure, res)
                    );
                    return observableThrowError(error);
                }));
    }
    operateRule(operation: string, rule: ReplicationRule): void {
        let title: string;
        let summary: string;
        let buttons: ConfirmationButtons;
        switch (operation) {
            case 'enable':
                title = 'REPLICATION.ENABLE_TITLE';
                summary = 'REPLICATION.ENABLE_SUMMARY';
                buttons = ConfirmationButtons.ENABLE_CANCEL;
                break;
            case 'disable':
                title = 'REPLICATION.DISABLE_TITLE';
                summary = 'REPLICATION.DISABLE_SUMMARY';
                buttons = ConfirmationButtons.DISABLE_CANCEL;
                break;

            default:
                return;
        }
        // Confirm
        const msg: ConfirmationMessage = new ConfirmationMessage(
            title,
            summary,
            rule.name,
            rule,
            ConfirmationTargets.REPLICATION,
            buttons
        );
        this.deletionConfirmDialog.open(msg);
    }
    clrLoad(state?: ClrDatagridStateInterface) {
        if (state && state.page) {
            this.pageSize = state.page.size;
        }
        this.loading = true;
        this.replicationService.getReplicationRulesResponse(
            this.ruleName,
            this.page,
            this.pageSize)
            .pipe(finalize(() => this.loading = false))
            .subscribe(response => {
              // job list hidden
              this.hideJobs.emit();
              // Get total count
              if (response.headers) {
                  let xHeader: string = response.headers.get("x-total-count");
                  if (xHeader) {
                      this.totalCount = parseInt(xHeader, 0);
                  }
              }
              this.rules = response.body as ReplicationRule[];
            }, error => {
              this.errorHandler.error(error);
            });
    }
    refreshRule() {
        this.page = 1;
        this.totalCount = 0;
        this.selectedRow = null;
        this.ruleName = "";
        this.clrLoad();
    }
    isHelmHub(srcRegistry: any): boolean {
      return srcRegistry && srcRegistry.type === HELM_HUB;
    }
}
