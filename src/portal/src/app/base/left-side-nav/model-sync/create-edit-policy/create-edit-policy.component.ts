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
import { Component, EventEmitter, Output, ViewChild } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { TranslateService } from '@ngx-translate/core';
import { forkJoin } from 'rxjs';
import { finalize } from 'rxjs/operators';
import { ModelSyncService } from '../../../../../../ng-swagger-gen/services/model-sync.service';
import { RegistryService } from '../../../../../../ng-swagger-gen/services/registry.service';
import { ProjectService } from '../../../../../../ng-swagger-gen/services/project.service';
import { ModelSyncPolicy } from '../../../../../../ng-swagger-gen/models/model-sync-policy';
import { ModelSyncPreview } from '../../../../../../ng-swagger-gen/models/model-sync-preview';
import { Registry } from '../../../../../../ng-swagger-gen/models/registry';
import { Project } from '../../../../../../ng-swagger-gen/models/project';
import { ErrorHandler } from '../../../../shared/units/error-handler';
import { InlineAlertComponent } from '../../../../shared/components/inline-alert/inline-alert.component';
import { cronRegex, formatSize } from '../../../../shared/units/utils';
import {
    formatFileFilters,
    parseFileFilters,
    TRIGGER_MANUAL,
    TRIGGER_SCHEDULED,
} from '../model-sync';

// registries are listed with a large page to fill the select
const LIST_PAGE_SIZE = 100;

@Component({
    selector: 'create-edit-model-sync-policy',
    templateUrl: './create-edit-policy.component.html',
    styleUrls: ['./create-edit-policy.component.scss'],
    standalone: false,
})
export class CreateEditModelSyncPolicyComponent {
    @Output() saved = new EventEmitter<void>();
    @Output() goToRegistry = new EventEmitter<void>();
    @ViewChild(InlineAlertComponent, { static: true })
    inlineAlert: InlineAlertComponent;

    opened: boolean = false;
    isEdit: boolean = false;
    policyId: number;
    saving: boolean = false;
    loadingOptions: boolean = false;

    form: FormGroup;
    registries: Registry[] = [];
    adapterTypes: string[] = [];
    projects: Project[] = [];

    // preview
    previewing: boolean = false;
    preview: ModelSyncPreview;
    previewError: string;
    showAllFiles: boolean = false;
    // the digest is computed by the backend for the filters sent with the
    // preview; a local re-match after a filter change cannot update it
    digestStale: boolean = false;

    triggerTypes = [TRIGGER_MANUAL, TRIGGER_SCHEDULED];

    constructor(
        private fb: FormBuilder,
        private modelSyncService: ModelSyncService,
        private registryService: RegistryService,
        private projectService: ProjectService,
        private errorHandler: ErrorHandler,
        private translate: TranslateService
    ) {
        this.form = this.fb.group({
            name: ['', [Validators.required, Validators.maxLength(256)]],
            description: [''],
            registry_id: [null, Validators.required],
            src_repository: [
                '',
                [
                    Validators.required,
                    Validators.pattern(
                        /^[A-Za-z0-9][A-Za-z0-9._-]*(\/[A-Za-z0-9][A-Za-z0-9._-]*)?$/
                    ),
                ],
            ],
            src_revision: [''],
            file_filters: [''],
            file_excludes: [''],
            dest_project_id: [null, Validators.required],
            dest_repository: [''],
            trigger_type: [TRIGGER_MANUAL, Validators.required],
            cron: [''],
            enabled: [true],
        });
        // invalidate the preview when the source changes
        ['registry_id', 'src_repository', 'src_revision'].forEach(key =>
            this.form.get(key).valueChanges.subscribe(() => this.resetPreview())
        );
        ['file_filters', 'file_excludes'].forEach(key =>
            this.form
                .get(key)
                .valueChanges.subscribe(() => this.recomputeMatches())
        );
    }

    // ---------- open / close ----------

    openCreate(): void {
        this.isEdit = false;
        this.policyId = null;
        this.form.reset({
            name: '',
            description: '',
            registry_id: null,
            src_repository: '',
            src_revision: '',
            file_filters: '',
            file_excludes: '',
            dest_project_id: null,
            dest_repository: '',
            trigger_type: TRIGGER_MANUAL,
            cron: '',
            enabled: true,
        });
        this.resetPreview();
        this.loadOptions();
        this.opened = true;
    }

    openEdit(policy: ModelSyncPolicy): void {
        this.isEdit = true;
        this.policyId = policy.id;
        this.form.reset({
            name: policy.name,
            description: policy.description || '',
            registry_id: policy.registry_id || policy.registry?.id || null,
            src_repository: policy.src_repository,
            src_revision: policy.src_revision || '',
            file_filters: formatFileFilters(policy.file_filters),
            file_excludes: formatFileFilters(policy.file_excludes),
            dest_project_id: policy.dest_project_id,
            dest_repository: policy.dest_repository || '',
            trigger_type: policy.trigger?.type || TRIGGER_MANUAL,
            cron: policy.trigger?.trigger_settings?.cron || '',
            enabled: policy.enabled,
        });
        this.resetPreview();
        this.loadOptions();
        this.opened = true;
    }

    close(): void {
        this.opened = false;
        this.inlineAlert.close();
    }

    onCancel(): void {
        if (this.form.dirty) {
            this.inlineAlert.showInlineConfirmation({
                message: 'ALERT.FORM_CHANGE_CONFIRMATION',
            });
        } else {
            this.close();
        }
    }

    confirmCancel(): void {
        this.close();
    }

    goRegistry(): void {
        this.close();
        this.goToRegistry.emit();
    }

    loadOptions(): void {
        this.loadingOptions = true;
        forkJoin([
            this.modelSyncService.listModelSyncAdapters(),
            this.registryService.listRegistries({ pageSize: LIST_PAGE_SIZE }),
            this.projectService.listProjects({
                pageSize: LIST_PAGE_SIZE,
                withDetail: false,
            }),
        ])
            .pipe(finalize(() => (this.loadingOptions = false)))
            .subscribe(
                ([types, registries, projects]) => {
                    this.adapterTypes = types || [];
                    this.registries = (registries || []).filter(r =>
                        this.adapterTypes.includes(r.type)
                    );
                    // proxy cache projects can't be destinations
                    this.projects = (projects || []).filter(
                        p => !p.registry_id
                    );
                },
                error => this.errorHandler.error(error)
            );
    }

    // ---------- helpers ----------

    get isScheduled(): boolean {
        return this.form.get('trigger_type').value === TRIGGER_SCHEDULED;
    }

    get srcRepository(): string {
        return (this.form.get('src_repository').value || '').trim();
    }

    get defaultDestRepository(): string {
        return this.srcRepository.replace(/^\/+|\/+$/g, '').toLowerCase();
    }

    get destProjectName(): string {
        const id = this.form.get('dest_project_id').value;
        const project = this.projects.find(p => p.project_id === +id);
        return project ? project.name : '';
    }

    get destinationPreview(): string {
        const repo =
            (this.form.get('dest_repository').value || '').trim() ||
            this.defaultDestRepository;
        return this.destProjectName && repo
            ? `${this.destProjectName}/${repo}`
            : '';
    }

    cronInvalid(): boolean {
        if (!this.isScheduled) {
            return false;
        }
        const cron = (this.form.get('cron').value || '').trim();
        return !cron || !cronRegex(cron);
    }

    cronShouldShowError(): boolean {
        const ctl = this.form.get('cron');
        return (ctl.touched || ctl.dirty) && this.cronInvalid();
    }

    fieldInvalid(name: string): boolean {
        const ctl = this.form.get(name);
        return ctl.invalid && (ctl.touched || ctl.dirty);
    }

    canSave(): boolean {
        return this.form.valid && !this.cronInvalid() && !this.saving;
    }

    canPreview(): boolean {
        return (
            !!this.form.get('registry_id').value &&
            this.form.get('src_repository').valid &&
            !this.previewing
        );
    }

    size(bytes: number): string {
        return formatSize(String(bytes || 0));
    }

    get displayedFiles() {
        if (!this.preview?.files) {
            return [];
        }
        return this.showAllFiles
            ? this.preview.files
            : this.preview.files.filter(f => f.matched);
    }

    // ---------- preview ----------

    resetPreview(): void {
        this.preview = null;
        this.previewError = null;
        this.showAllFiles = false;
        this.digestStale = false;
    }

    // recompute matched files locally when the filters change so the user
    // doesn't need to re-resolve the source
    recomputeMatches(): void {
        if (!this.preview?.files) {
            return;
        }
        const includes = parseFileFilters(this.form.get('file_filters').value);
        const excludes = parseFileFilters(this.form.get('file_excludes').value);
        let matchedFiles = 0;
        let matchedSize = 0;
        this.digestStale = true;
        this.preview.files.forEach(f => {
            f.matched = fileSelected(includes, excludes, f.path);
            if (f.matched) {
                matchedFiles++;
                matchedSize += f.size || 0;
            }
        });
        this.preview.matched_files = matchedFiles;
        this.preview.matched_size = matchedSize;
    }

    doPreview(): void {
        if (!this.canPreview()) {
            return;
        }
        this.previewing = true;
        this.previewError = null;
        this.modelSyncService
            .previewModelSync({
                preview: {
                    registry_id: +this.form.get('registry_id').value,
                    src_repository: this.srcRepository,
                    src_revision: (
                        this.form.get('src_revision').value || ''
                    ).trim(),
                    file_filters: parseFileFilters(
                        this.form.get('file_filters').value
                    ),
                    file_excludes: parseFileFilters(
                        this.form.get('file_excludes').value
                    ),
                },
            })
            .pipe(finalize(() => (this.previewing = false)))
            .subscribe(
                res => {
                    this.preview = res;
                    this.showAllFiles = false;
                    this.digestStale = false;
                },
                error => {
                    this.preview = null;
                    this.previewError =
                        error?.error?.errors?.[0]?.message ||
                        error?.message ||
                        'MODEL_SYNC.PREVIEW_FAILED';
                }
            );
    }

    // ---------- save ----------

    toPolicy(): ModelSyncPolicy {
        const v = this.form.value;
        const policy: ModelSyncPolicy = {
            name: (v.name || '').trim(),
            description: v.description || '',
            registry_id: +v.registry_id,
            src_repository: this.srcRepository,
            src_revision: (v.src_revision || '').trim(),
            file_filters: parseFileFilters(v.file_filters),
            file_excludes: parseFileFilters(v.file_excludes),
            dest_project_id: +v.dest_project_id,
            dest_repository: (v.dest_repository || '').trim(),
            enabled: !!v.enabled,
            trigger: { type: v.trigger_type },
        };
        if (v.trigger_type === TRIGGER_SCHEDULED) {
            policy.trigger.trigger_settings = { cron: (v.cron || '').trim() };
        }
        return policy;
    }

    onSave(): void {
        if (!this.canSave()) {
            return;
        }
        this.saving = true;
        const policy = this.toPolicy();
        const request = this.isEdit
            ? this.modelSyncService.updateModelSyncPolicy({
                  id: this.policyId,
                  policy,
              })
            : this.modelSyncService.createModelSyncPolicy({ policy });
        request.pipe(finalize(() => (this.saving = false))).subscribe(
            () => {
                this.translate
                    .get(
                        this.isEdit
                            ? 'MODEL_SYNC.UPDATED_SUCCESS'
                            : 'MODEL_SYNC.CREATED_SUCCESS'
                    )
                    .subscribe(res => this.errorHandler.info(res));
                this.close();
                this.saved.emit();
            },
            error => this.inlineAlert.showInlineError(error)
        );
    }
}

// fileSelected mirrors model.FileFilter.Match: any include (or none) and no exclude.
export function fileSelected(
    includes: string[],
    excludes: string[],
    path: string
): boolean {
    if (includes.length > 0 && !matchesAny(includes, path)) {
        return false;
    }
    return !matchesAny(excludes, path);
}

// matchesAny mirrors the backend filter semantics closely enough for the
// preview: a pattern matches the full path, or the base name when the pattern
// has no path separator.
export function matchesAny(patterns: string[], path: string): boolean {
    const base = path.split('/').pop();
    return patterns.some(p => {
        const re = globToRegExp(p);
        return re.test(path) || (!p.includes('/') && re.test(base));
    });
}

export function globToRegExp(glob: string): RegExp {
    let re = '';
    for (let i = 0; i < glob.length; i++) {
        const c = glob[i];
        if (c === '*') {
            if (glob[i + 1] === '*') {
                re += '.*';
                i++;
                if (glob[i + 1] === '/') {
                    i++;
                }
            } else {
                re += '[^/]*';
            }
        } else if (c === '?') {
            re += '[^/]';
        } else if ('.+^${}()|[]\\'.includes(c)) {
            re += '\\' + c;
        } else {
            re += c;
        }
    }
    return new RegExp(`^${re}$`);
}
