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
import { TestBed } from '@angular/core/testing';
import { HttpHeaders, HttpResponse } from '@angular/common/http';
import { of } from 'rxjs';
import { ReplicationService } from '../../../../../ng-swagger-gen/services/replication.service';
import { RegistryService } from '../../../../../ng-swagger-gen/services/registry.service';
import { ModelSyncService } from './model-sync.service';
import {
    destRepository,
    fromReplicationPolicy,
    ModelSyncPolicy,
    toReplicationPolicy,
} from './model-sync';

describe('ModelSyncService', () => {
    const policy: ModelSyncPolicy = {
        id: 1,
        name: 'smollm',
        description: 'd',
        enabled: true,
        registry_id: 2,
        src_repository: 'HuggingFaceTB/SmolLM2-135M',
        src_revision: 'main',
        file_filters: ['*.json'],
        file_excludes: ['onnx/**'],
        dest_project_name: 'models',
        dest_repository: 'smollm2',
        trigger: {
            type: 'scheduled',
            trigger_settings: { cron: '0 0 * * * *' },
        },
    };
    const rule = {
        ...toReplicationPolicy(policy),
        id: 1,
        src_registry: { id: 2, name: 'hf', type: 'huggingface' },
    };
    let replication: jasmine.SpyObj<ReplicationService>;
    let service: ModelSyncService;

    beforeEach(() => {
        replication = jasmine.createSpyObj('ReplicationService', [
            'listReplicationPoliciesResponse',
            'startReplication',
        ]);
        replication.listReplicationPoliciesResponse.and.returnValue(
            of(
                new HttpResponse({
                    body: [rule],
                    headers: new HttpHeaders({ 'X-Total-Count': '1' }),
                })
            ) as any
        );
        const registry = {
            listRegistryProviderInfos: () =>
                of({
                    huggingface: { supported_preview: true },
                    'docker-hub': {},
                }),
            listRegistries: () =>
                of([
                    { id: 2, type: 'huggingface' },
                    { id: 3, type: 'docker-hub' },
                    { id: 4, type: 'huggingface' },
                ]),
        };
        TestBed.configureTestingModule({
            providers: [
                ModelSyncService,
                { provide: ReplicationService, useValue: replication },
                { provide: RegistryService, useValue: registry },
            ],
        });
        service = TestBed.inject(ModelSyncService);
    });

    it('maps a policy to a replication rule and back', () => {
        const r = toReplicationPolicy(policy);
        expect(r.src_registry).toEqual({ id: 2 });
        expect(r.dest_registry).toBeUndefined();
        expect(r.dest_namespace).toEqual('models');
        expect(r.dest_namespace_replace_count).toEqual(-1);
        expect(r.dest_repository).toEqual('smollm2');
        expect(r.revision_tag).toBeTrue();
        expect(r.override).toBeTrue();
        expect(r.filters).toEqual([
            { type: 'name', value: 'HuggingFaceTB/SmolLM2-135M' },
            { type: 'tag', value: 'main', decoration: 'matches' },
            {
                type: 'file',
                value: { include: ['*.json'], exclude: ['onnx/**'] },
            },
        ]);
        const back = fromReplicationPolicy({
            ...r,
            id: 1,
            src_registry: rule.src_registry,
        });
        expect(back).toEqual(jasmine.objectContaining(policy));
    });

    it('always sends a revision so a rule never copies every branch', () => {
        const r = toReplicationPolicy({
            ...policy,
            src_revision: '',
            file_filters: [],
            file_excludes: [],
        });
        expect(r.filters).toEqual([
            { type: 'name', value: 'HuggingFaceTB/SmolLM2-135M' },
            { type: 'tag', value: 'main', decoration: 'matches' },
        ]);
    });

    it('derives the destination like the replication rewrite', () => {
        expect(destRepository(policy)).toEqual('models/smollm2');
        expect(destRepository({ ...policy, dest_repository: '' })).toEqual(
            'models/smollm2-135m'
        );
        expect(
            destRepository({ dest_project_name: 'm', src_repository: 'gpt2' })
        ).toEqual('m/gpt2');
    });

    it('lists only rules of model sources', done => {
        service
            .listPolicies({ page: 1, pageSize: 5 }, 'smol')
            .subscribe(page => {
                expect(page.total).toEqual(1);
                expect(page.items[0].src_repository).toEqual(
                    'HuggingFaceTB/SmolLM2-135M'
                );
                const params =
                    replication.listReplicationPoliciesResponse.calls.mostRecent()
                        .args[0];
                expect(decodeURIComponent(params.q)).toEqual(
                    'src_registry_id={2 4},name=~smol'
                );
                done();
            });
    });

    it('lists nothing without a model source registry', done => {
        TestBed.inject(RegistryService).listRegistries = (() =>
            of([{ id: 3, type: 'docker-hub' }])) as any;
        service.listPolicies({ page: 1, pageSize: 5 }).subscribe(page => {
            expect(page).toEqual({ items: [], total: 0 });
            expect(
                replication.listReplicationPoliciesResponse
            ).not.toHaveBeenCalled();
            done();
        });
    });

    it('starts a replication execution', () => {
        replication.startReplication.and.returnValue(of(null));
        service.start(1).subscribe();
        expect(replication.startReplication).toHaveBeenCalledWith({
            execution: { policy_id: 1 },
        });
    });
});
