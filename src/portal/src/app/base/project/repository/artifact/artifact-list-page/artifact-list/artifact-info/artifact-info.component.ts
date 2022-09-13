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
import { Component, OnInit, ViewChild } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { TranslateService } from '@ngx-translate/core';
import { RepositoryService } from 'ng-swagger-gen/services/repository.service';
import { ConfirmationMessage } from 'src/app/base/global-confirmation-dialog/confirmation-message';
import { ConfirmationAcknowledgement } from 'src/app/base/global-confirmation-dialog/confirmation-state-message';
import { Project } from 'src/app/base/project/project';
import { ConfirmationDialogComponent } from 'src/app/shared/components/confirmation-dialog/confirmation-dialog.component';
import {
    ConfirmationState,
    ConfirmationTargets,
} from 'src/app/shared/entities/shared.const';
import { ErrorHandler } from 'src/app/shared/units/error-handler/error-handler';
import { dbEncodeURIComponent } from 'src/app/shared/units/utils';
import { finalize } from 'rxjs/operators';

@Component({
    selector: 'artifact-info',
    templateUrl: './artifact-info.component.html',
    styleUrls: ['./artifact-info.component.scss'],
})
export class ArtifactInfoComponent implements OnInit {
    projectName: string;
    repoName: string;
    hasProjectAdminRole: boolean = false;
    onSaving: boolean = false;
    loading: boolean = false;
    editing: boolean = false;
    imageInfo: string;
    orgImageInfo: string;
    @ViewChild('confirmationDialog')
    confirmationDlg: ConfirmationDialogComponent;

    constructor(
        private errorHandler: ErrorHandler,
        private repositoryService: RepositoryService,
        private translate: TranslateService,
        private activatedRoute: ActivatedRoute
    ) {}

    ngOnInit(): void {
        this.repoName = this.activatedRoute.snapshot?.parent?.params['repo'];
        let resolverData = this.activatedRoute.snapshot?.parent?.parent?.data;
        if (resolverData) {
            this.projectName = (<Project>resolverData['projectResolver']).name;
            this.hasProjectAdminRole = (<Project>(
                resolverData['projectResolver']
            )).has_project_admin_role;
        }
        this.retrieve();
    }

    retrieve() {
        let params: RepositoryService.GetRepositoryParams = {
            projectName: this.projectName,
            repositoryName: dbEncodeURIComponent(this.repoName),
        };
        this.loading = true;
        this.repositoryService
            .getRepository(params)
            .pipe(finalize(() => (this.loading = false)))
            .subscribe(
                response => {
                    this.orgImageInfo = response.description;
                    this.imageInfo = response.description;
                },
                error => this.errorHandler.error(error)
            );
    }

    refresh() {
        this.retrieve();
    }

    hasChanges() {
        return this.imageInfo !== this.orgImageInfo;
    }

    reset(): void {
        this.imageInfo = this.orgImageInfo;
    }

    editInfo() {
        this.editing = true;
    }

    saveInfo() {
        if (!this.hasChanges()) {
            return;
        }
        this.onSaving = true;
        let params: RepositoryService.UpdateRepositoryParams = {
            repositoryName: dbEncodeURIComponent(this.repoName),
            repository: { description: this.imageInfo },
            projectName: this.projectName,
        };
        this.repositoryService.updateRepository(params).subscribe(
            () => {
                this.onSaving = false;
                this.translate
                    .get('CONFIG.SAVE_SUCCESS')
                    .subscribe((res: string) => {
                        this.errorHandler.info(res);
                    });
                this.editing = false;
                this.refresh();
            },
            error => {
                this.onSaving = false;
                this.errorHandler.error(error);
            }
        );
    }

    cancelInfo() {
        let msg = new ConfirmationMessage(
            'CONFIG.CONFIRM_TITLE',
            'CONFIG.CONFIRM_SUMMARY',
            '',
            {},
            ConfirmationTargets.CONFIG
        );
        this.confirmationDlg.open(msg);
    }

    confirmCancel(ack: ConfirmationAcknowledgement): void {
        this.editing = false;
        if (
            ack &&
            ack.source === ConfirmationTargets.CONFIG &&
            ack.state === ConfirmationState.CONFIRMED
        ) {
            this.reset();
        }
    }
}
