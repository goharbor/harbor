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

import { Component, EventEmitter, Input, Output } from '@angular/core';
import { finalize, switchMap } from 'rxjs/operators';
import {
    ModelImportExecutionRequest,
    ModelImportPolicy,
    ModelImportService,
    ModelImportValidation,
} from './model-import.service';
import { ErrorHandler } from '../../../../shared/units/error-handler';

@Component({
    selector: 'hbr-model-import',
    templateUrl: './model-import.component.html',
    styleUrls: ['./model-import.component.scss'],
    standalone: false,
})
export class ModelImportComponent {
    @Input() projectName: string;
    @Input() opened = false;
    @Output() openedChange = new EventEmitter<boolean>();
    @Output() imported = new EventEmitter<void>();

    policyName = '';
    description = '';
    hfRepoId = '';
    hfRevision = 'main';
    hfSubpath = '';
    hfToken = '';
    targetRepository = '';
    targetTag = 'latest';
    scheduleEnabled = false;
    cron = '';
    runNow = true;

    validation: ModelImportValidation;
    validating = false;
    saving = false;

    constructor(
        private modelImportService: ModelImportService,
        private errorHandler: ErrorHandler
    ) {}

    close(): void {
        this.opened = false;
        this.openedChange.emit(false);
    }

    validate(): void {
        if (!this.hfRepoId) {
            return;
        }
        this.validating = true;
        this.modelImportService
            .validateSource(this.projectName, {
                hf_repo_id: this.hfRepoId,
                hf_revision: this.hfRevision || 'main',
                hf_subpath: this.hfSubpath,
                hf_token: this.hfToken,
            })
            .pipe(finalize(() => (this.validating = false)))
            .subscribe({
                next: result => {
                    this.validation = result;
                    this.applyDefaultsFromSource();
                },
                error: err => this.errorHandler.error(err),
            });
    }

    create(): void {
        this.saving = true;
        const request$ = this.scheduleEnabled
            ? this.createScheduledPolicy()
            : this.modelImportService.runImport(
                  this.projectName,
                  this.buildExecutionRequest()
              );
        request$.pipe(finalize(() => (this.saving = false))).subscribe({
            next: () => {
                this.imported.emit();
                this.close();
            },
            error: err => this.errorHandler.error(err),
        });
    }

    private createScheduledPolicy() {
        const policy = this.buildPolicy();
        const create$ = this.modelImportService.createPolicy(
            this.projectName,
            policy
        );
        return this.runNow
            ? create$.pipe(
                  switchMap(() =>
                      this.modelImportService.runPolicy(
                          this.projectName,
                          policy.name
                      )
                  )
              )
            : create$;
    }

    private applyDefaultsFromSource(): void {
        const name = this.hfRepoId.split('/').pop() || this.hfRepoId;
        if (!this.policyName) {
            this.policyName = name.replace(/[^A-Za-z0-9._-]/g, '-');
        }
        if (!this.targetRepository) {
            this.targetRepository = `${this.projectName}/${name}`;
        }
    }

    private buildPolicy(): ModelImportPolicy {
        const trigger = this.scheduleEnabled
            ? {
                  type: 'scheduled',
                  trigger_setting: {
                      cron: this.cron,
                  },
              }
            : {
                  type: 'manual',
              };
        return {
            name: this.policyName,
            description: this.description,
            target_repository: this.normalizedTargetRepository(),
            target_tag: this.targetTag || 'latest',
            hf_repo_id: this.hfRepoId,
            hf_revision: this.hfRevision || 'main',
            hf_subpath: this.hfSubpath,
            hf_token: this.hfToken,
            enabled: true,
            trigger: JSON.stringify(trigger),
        };
    }

    private buildExecutionRequest(): ModelImportExecutionRequest {
        return {
            target_repository: this.normalizedTargetRepository(),
            target_tag: this.targetTag || 'latest',
            hf_repo_id: this.hfRepoId,
            hf_revision: this.hfRevision || 'main',
            hf_subpath: this.hfSubpath,
            hf_token: this.hfToken,
        };
    }

    private normalizedTargetRepository(): string {
        const repository = (this.targetRepository || '').trim().replace(/^\/+|\/+$/g, '');
        if (!repository || repository.includes('/')) {
            return repository;
        }
        return `${this.projectName}/${repository}`;
    }
}
