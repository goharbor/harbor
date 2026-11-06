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
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { of, throwError } from 'rxjs';
import { SharedTestingModule } from '../../../../shared/shared.module';
import { ModelSyncService } from '../../../../../../ng-swagger-gen/services/model-sync.service';
import { RegistryService } from '../../../../../../ng-swagger-gen/services/registry.service';
import { ProjectService } from '../../../../../../ng-swagger-gen/services/project.service';
import {
    CreateEditModelSyncPolicyComponent,
    fileSelected,
    globToRegExp,
    matchesAny,
} from './create-edit-policy.component';
import { ErrorHandler } from '../../../../shared/units/error-handler';

describe('CreateEditModelSyncPolicyComponent', () => {
    let component: CreateEditModelSyncPolicyComponent;
    let fixture: ComponentFixture<CreateEditModelSyncPolicyComponent>;

    const preview = {
        repository: 'Qwen/Qwen3-8B',
        ref: 'main',
        revision: '71034c5d8bde858ff824298bdedc65515b97d2b9',
        source_url: 'https://huggingface.co/Qwen/Qwen3-8B/tree/x',
        tags: ['main', 'sha-71034c5d8bde'],
        revision_tag: 'sha-71034c5d8bde',
        digest: 'sha256:2dc8fb68cff37ff1f1ae1bcef28039072cb9a012d4e7c7c817cee9d3d1b3b18e',
        metadata: { license: 'apache-2.0' },
        total_files: 2,
        total_size: 110,
        matched_files: 1,
        matched_size: 100,
        files: [
            {
                path: 'model.safetensors',
                size: 100,
                type: 'weight',
                matched: true,
            },
            { path: 'README.md', size: 10, type: 'doc', matched: false },
        ],
    };
    const mockModelSync = {
        listModelSyncAdapters: () => of(['huggingface']),
        previewModelSync: jasmine
            .createSpy('previewModelSync')
            .and.returnValue(of(preview)),
        createModelSyncPolicy: jasmine
            .createSpy('createModelSyncPolicy')
            .and.returnValue(of(null)),
        updateModelSyncPolicy: jasmine
            .createSpy('updateModelSyncPolicy')
            .and.returnValue(of(null)),
    };
    const mockRegistry = {
        listRegistries: () =>
            of([
                {
                    id: 2,
                    name: 'hf',
                    type: 'huggingface',
                    url: 'https://huggingface.co',
                },
                {
                    id: 3,
                    name: 'docker',
                    type: 'docker-hub',
                    url: 'https://hub.docker.com',
                },
            ]),
    };
    const mockProject = {
        listProjects: () =>
            of([
                { project_id: 1, name: 'library' },
                { project_id: 2, name: 'proxy', registry_id: 5 },
            ]),
    };
    const mockErrorHandler = { error: () => {}, info: () => {} };

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
            imports: [SharedTestingModule],
            declarations: [CreateEditModelSyncPolicyComponent],
            providers: [
                { provide: ModelSyncService, useValue: mockModelSync },
                { provide: RegistryService, useValue: mockRegistry },
                { provide: ProjectService, useValue: mockProject },
                { provide: ErrorHandler, useValue: mockErrorHandler },
            ],
        }).compileComponents();
        fixture = TestBed.createComponent(CreateEditModelSyncPolicyComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should open for creation and load filtered options', () => {
        component.openCreate();
        expect(component.opened).toBeTrue();
        expect(component.isEdit).toBeFalse();
        // only registries with a model adapter, only non proxy projects
        expect(component.registries.map(r => r.id)).toEqual([2]);
        expect(component.projects.map(p => p.project_id)).toEqual([1]);
        expect(component.canSave()).toBeFalse();
    });

    it('should validate the form and build the policy', () => {
        component.openCreate();
        component.form.patchValue({
            name: 'qwen',
            registry_id: 2,
            src_repository: 'Qwen/Qwen3-8B',
            file_filters: '*.safetensors, *.json',
            file_excludes: 'onnx/**\n *.bin',
            dest_project_id: 1,
            trigger_type: 'scheduled',
            cron: '0 0 * * * *',
        });
        expect(component.canSave()).toBeTrue();
        expect(component.defaultDestRepository).toEqual('qwen/qwen3-8b');
        expect(component.destinationPreview).toEqual('library/qwen/qwen3-8b');
        const policy = component.toPolicy();
        expect(policy.file_filters).toEqual(['*.safetensors', '*.json']);
        expect(policy.file_excludes).toEqual(['onnx/**', '*.bin']);
        expect(policy.trigger.trigger_settings.cron).toEqual('0 0 * * * *');

        component.form.patchValue({ cron: 'bad' });
        expect(component.cronInvalid()).toBeTrue();
        expect(component.canSave()).toBeFalse();
        component.form.patchValue({ trigger_type: 'manual' });
        expect(component.cronInvalid()).toBeFalse();
        expect(component.toPolicy().trigger.trigger_settings).toBeUndefined();
    });

    it('should open for editing', () => {
        component.openEdit({
            id: 7,
            name: 'p',
            registry_id: 2,
            src_repository: 'a/b',
            file_filters: ['*.bin'],
            file_excludes: ['a/**', '*.md'],
            dest_project_id: 1,
            trigger: { type: 'manual' },
            enabled: false,
        });
        expect(component.isEdit).toBeTrue();
        expect(component.policyId).toEqual(7);
        expect(component.form.get('file_filters').value).toEqual('*.bin');
        expect(component.form.get('file_excludes').value).toEqual('a/**, *.md');
        expect(component.form.get('enabled').value).toBeFalse();
    });

    it('should preview and recompute matches locally', () => {
        component.openCreate();
        component.form.patchValue({
            registry_id: 2,
            src_repository: 'Qwen/Qwen3-8B',
            file_filters: '*.safetensors',
        });
        expect(component.canPreview()).toBeTrue();
        component.doPreview();
        expect(mockModelSync.previewModelSync).toHaveBeenCalledWith({
            preview: jasmine.objectContaining({
                file_filters: ['*.safetensors'],
                file_excludes: [],
            }),
        });
        expect(component.preview).toBeTruthy();
        expect(component.digestStale).toBeFalse();
        expect(component.displayedFiles.length).toEqual(1);
        component.showAllFiles = true;
        expect(component.displayedFiles.length).toEqual(2);

        component.form.patchValue({ file_filters: '' });
        expect(component.preview.matched_files).toEqual(2);
        component.form.patchValue({ file_filters: 'README.md' });
        expect(component.preview.matched_files).toEqual(1);
        expect(component.preview.matched_size).toEqual(10);
        expect(component.digestStale).toBeTrue();
        component.form.patchValue({
            file_filters: '',
            file_excludes: '*.md',
        });
        expect(component.preview.matched_files).toEqual(1);
        expect(component.preview.matched_size).toEqual(100);

        // changing the source invalidates the preview
        component.form.patchValue({ src_revision: 'v2' });
        expect(component.preview).toBeNull();
    });

    it('should surface preview errors', () => {
        mockModelSync.previewModelSync.and.returnValue(
            throwError(() => ({
                error: { errors: [{ message: 'not found' }] },
            }))
        );
        component.openCreate();
        component.form.patchValue({ registry_id: 2, src_repository: 'a/b' });
        component.doPreview();
        expect(component.preview).toBeNull();
        expect(component.previewError).toEqual('not found');
        mockModelSync.previewModelSync.and.returnValue(of(preview));
    });

    it('should save', () => {
        let saved = false;
        component.saved.subscribe(() => (saved = true));
        component.openCreate();
        component.form.patchValue({
            name: 'qwen',
            registry_id: 2,
            src_repository: 'Qwen/Qwen3-8B',
            dest_project_id: 1,
        });
        component.onSave();
        expect(mockModelSync.createModelSyncPolicy).toHaveBeenCalled();
        expect(saved).toBeTrue();
        expect(component.opened).toBeFalse();
    });

    it('should match globs like the backend', () => {
        expect(globToRegExp('*.json').test('config.json')).toBeTrue();
        expect(globToRegExp('*.json').test('sub/config.json')).toBeFalse();
        expect(
            globToRegExp('**/*.json').test('sub/dir/config.json')
        ).toBeTrue();
        expect(matchesAny(['*.json'], 'sub/config.json')).toBeTrue();
        expect(matchesAny(['sub/*.json'], 'sub/dir/config.json')).toBeFalse();
        expect(matchesAny(['README.md'], 'readme.md')).toBeFalse();
    });

    it('should apply includes then excludes like model.FileFilter', () => {
        expect(fileSelected([], [], 'onnx/model.onnx')).toBeTrue();
        expect(fileSelected(['*.json'], [], 'onnx/config.json')).toBeTrue();
        expect(
            fileSelected(['*.json'], ['onnx/**'], 'onnx/config.json')
        ).toBeFalse();
        expect(fileSelected([], ['*.bin'], 'pytorch_model.bin')).toBeFalse();
        expect(fileSelected(['*.json'], [], 'model.safetensors')).toBeFalse();
    });
});
