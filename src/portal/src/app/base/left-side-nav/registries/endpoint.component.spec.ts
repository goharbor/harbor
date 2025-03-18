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
import { NO_ERRORS_SCHEMA } from '@angular/core';
import { EndpointComponent } from './endpoint.component';
import { CreateEditEndpointComponent } from './create-edit-endpoint/create-edit-endpoint.component';
import { ErrorHandler } from '../../../shared/units/error-handler';
import { OperationService } from '../../../shared/components/operation/operation.service';
import { click } from '../../../shared/units/utils';
import { of } from 'rxjs';
import { HttpClient, HttpHeaders, HttpResponse } from '@angular/common/http';
import { AppConfigService } from '../../../services/app-config.service';
import { SharedTestingModule } from '../../../shared/shared.module';
import {
    ADAPTERS_MAP,
    EndpointService,
} from '../../../shared/services/endpoint.service';
import { delay } from 'rxjs/operators';
import { RegistryService } from '../../../../../ng-swagger-gen/services/registry.service';
import { Registry } from '../../../../../ng-swagger-gen/models/registry';

describe('EndpointComponent (inline template)', () => {
    let adapterInfoMockData = {
        'ali-acr': {
            endpoint_pattern: {
                endpoint_type: 'EndpointPatternTypeList',
                endpoints: [
                    {
                        key: 'cn-hangzhou',
                        value: 'https://registry.cn-hangzhou.aliyuncs.com',
                    },
                    {
                        key: 'cn-shanghai',
                        value: 'https://registry.cn-shanghai.aliyuncs.com',
                    },
                    {
                        key: 'cn-qingdao',
                        value: 'https://registry.cn-qingdao.aliyuncs.com',
                    },
                    {
                        key: 'cn-beijing',
                        value: 'https://registry.cn-beijing.aliyuncs.com',
                    },
                    {
                        key: 'cn-zhangjiakou',
                        value: 'https://registry.cn-zhangjiakou.aliyuncs.com',
                    },
                    {
                        key: 'cn-huhehaote',
                        value: 'https://registry.cn-huhehaote.aliyuncs.com',
                    },
                    {
                        key: 'cn-shenzhen',
                        value: 'https://registry.cn-shenzhen.aliyuncs.com',
                    },
                    {
                        key: 'cn-chengdu',
                        value: 'https://registry.cn-chengdu.aliyuncs.com',
                    },
                    {
                        key: 'cn-hongkong',
                        value: 'https://registry.cn-hongkong.aliyuncs.com',
                    },
                    {
                        key: 'ap-southeast-1',
                        value: 'https://registry.ap-southeast-1.aliyuncs.com',
                    },
                    {
                        key: 'ap-southeast-2',
                        value: 'https://registry.ap-southeast-2.aliyuncs.com',
                    },
                    {
                        key: 'ap-southeast-3',
                        value: 'https://registry.ap-southeast-3.aliyuncs.com',
                    },
                    {
                        key: 'ap-southeast-5',
                        value: 'https://registry.ap-southeast-5.aliyuncs.com',
                    },
                    {
                        key: 'ap-northeast-1',
                        value: 'https://registry.ap-northeast-1.aliyuncs.com',
                    },
                    {
                        key: 'ap-south-1',
                        value: 'https://registry.ap-south-1.aliyuncs.com',
                    },
                    {
                        key: 'eu-central-1',
                        value: 'https://registry.eu-central-1.aliyuncs.com',
                    },
                    {
                        key: 'eu-west-1',
                        value: 'https://registry.eu-west-1.aliyuncs.com',
                    },
                    {
                        key: 'us-west-1',
                        value: 'https://registry.us-west-1.aliyuncs.com',
                    },
                    {
                        key: 'us-east-1',
                        value: 'https://registry.us-east-1.aliyuncs.com',
                    },
                    {
                        key: 'me-east-1',
                        value: 'https://registry.me-east-1.aliyuncs.com',
                    },
                    {
                        key: 'cn-hangzhou-vpc',
                        value: 'https://registry-vpc.cn-hangzhou.aliyuncs.com',
                    },
                    {
                        key: 'cn-shanghai-vpc',
                        value: 'https://registry-vpc.cn-shanghai.aliyuncs.com',
                    },
                    {
                        key: 'cn-qingdao-vpc',
                        value: 'https://registry-vpc.cn-qingdao.aliyuncs.com',
                    },
                    {
                        key: 'cn-beijing-vpc',
                        value: 'https://registry-vpc.cn-beijing.aliyuncs.com',
                    },
                    {
                        key: 'cn-zhangjiakou-vpc',
                        value: 'https://registry-vpc.cn-zhangjiakou.aliyuncs.com',
                    },
                    {
                        key: 'cn-huhehaote-vpc',
                        value: 'https://registry-vpc.cn-huhehaote.aliyuncs.com',
                    },
                    {
                        key: 'cn-shenzhen-vpc',
                        value: 'https://registry-vpc.cn-shenzhen.aliyuncs.com',
                    },
                    {
                        key: 'cn-chengdu-vpc',
                        value: 'https://registry-vpc.cn-chengdu.aliyuncs.com',
                    },
                    {
                        key: 'cn-hongkong-vpc',
                        value: 'https://registry-vpc.cn-hongkong.aliyuncs.com',
                    },
                    {
                        key: 'ap-southeast-1-vpc',
                        value: 'https://registry-vpc.ap-southeast-1.aliyuncs.com',
                    },
                    {
                        key: 'ap-southeast-2-vpc',
                        value: 'https://registry-vpc.ap-southeast-2.aliyuncs.com',
                    },
                    {
                        key: 'ap-southeast-3-vpc',
                        value: 'https://registry-vpc.ap-southeast-3.aliyuncs.com',
                    },
                    {
                        key: 'ap-southeast-5-vpc',
                        value: 'https://registry-vpc.ap-southeast-5.aliyuncs.com',
                    },
                    {
                        key: 'ap-northeast-1-vpc',
                        value: 'https://registry-vpc.ap-northeast-1.aliyuncs.com',
                    },
                    {
                        key: 'ap-south-1-vpc',
                        value: 'https://registry-vpc.ap-south-1.aliyuncs.com',
                    },
                    {
                        key: 'eu-central-1-vpc',
                        value: 'https://registry-vpc.eu-central-1.aliyuncs.com',
                    },
                    {
                        key: 'eu-west-1-vpc',
                        value: 'https://registry-vpc.eu-west-1.aliyuncs.com',
                    },
                    {
                        key: 'us-west-1-vpc',
                        value: 'https://registry-vpc.us-west-1.aliyuncs.com',
                    },
                    {
                        key: 'us-east-1-vpc',
                        value: 'https://registry-vpc.us-east-1.aliyuncs.com',
                    },
                    {
                        key: 'me-east-1-vpc',
                        value: 'https://registry-vpc.me-east-1.aliyuncs.com',
                    },
                    {
                        key: 'cn-hangzhou-internal',
                        value: 'https://registry-internal.cn-hangzhou.aliyuncs.com',
                    },
                    {
                        key: 'cn-shanghai-internal',
                        value: 'https://registry-internal.cn-shanghai.aliyuncs.com',
                    },
                    {
                        key: 'cn-qingdao-internal',
                        value: 'https://registry-internal.cn-qingdao.aliyuncs.com',
                    },
                    {
                        key: 'cn-beijing-internal',
                        value: 'https://registry-internal.cn-beijing.aliyuncs.com',
                    },
                    {
                        key: 'cn-zhangjiakou-internal',
                        value: 'https://registry-internal.cn-zhangjiakou.aliyuncs.com',
                    },
                    {
                        key: 'cn-huhehaote-internal',
                        value: 'https://registry-internal.cn-huhehaote.aliyuncs.com',
                    },
                    {
                        key: 'cn-shenzhen-internal',
                        value: 'https://registry-internal.cn-shenzhen.aliyuncs.com',
                    },
                    {
                        key: 'cn-chengdu-internal',
                        value: 'https://registry-internal.cn-chengdu.aliyuncs.com',
                    },
                    {
                        key: 'cn-hongkong-internal',
                        value: 'https://registry-internal.cn-hongkong.aliyuncs.com',
                    },
                    {
                        key: 'ap-southeast-1-internal',
                        value: 'https://registry-internal.ap-southeast-1.aliyuncs.com',
                    },
                    {
                        key: 'ap-southeast-2-internal',
                        value: 'https://registry-internal.ap-southeast-2.aliyuncs.com',
                    },
                    {
                        key: 'ap-southeast-3-internal',
                        value: 'https://registry-internal.ap-southeast-3.aliyuncs.com',
                    },
                    {
                        key: 'ap-southeast-5-internal',
                        value: 'https://registry-internal.ap-southeast-5.aliyuncs.com',
                    },
                    {
                        key: 'ap-northeast-1-internal',
                        value: 'https://registry-internal.ap-northeast-1.aliyuncs.com',
                    },
                    {
                        key: 'ap-south-1-internal',
                        value: 'https://registry-internal.ap-south-1.aliyuncs.com',
                    },
                    {
                        key: 'eu-central-1-internal',
                        value: 'https://registry-internal.eu-central-1.aliyuncs.com',
                    },
                    {
                        key: 'eu-west-1-internal',
                        value: 'https://registry-internal.eu-west-1.aliyuncs.com',
                    },
                    {
                        key: 'us-west-1-internal',
                        value: 'https://registry-internal.us-west-1.aliyuncs.com',
                    },
                    {
                        key: 'us-east-1-internal',
                        value: 'https://registry-internal.us-east-1.aliyuncs.com',
                    },
                    {
                        key: 'me-east-1-internal',
                        value: 'https://registry-internal.me-east-1.aliyuncs.com',
                    },
                    {
                        key: 'cn-hangzhou-ee-vpc',
                        value: `https://instanceName-registry-vpc.cn-hangzhou.cr.aliyuncs.com`,
                    },
                    {
                        key: 'cn-shanghai-ee-vpc',
                        value: 'https://instanceName-registry-vpc.cn-shanghai.cr.aliyuncs.com',
                    },
                    {
                        key: 'cn-qingdao-ee-vpc',
                        value: 'https://instanceName-registry-vpc.cn-qingdao.cr.aliyuncs.com',
                    },
                    {
                        key: 'cn-beijing-ee-vpc',
                        value: 'https://instanceName-registry-vpc.cn-beijing.cr.aliyuncs.com',
                    },
                    {
                        key: 'cn-zhangjiakou-ee-vpc',
                        value: 'https://instanceName-registry-vpc.cn-zhangjiakou.cr.aliyuncs.com',
                    },
                    {
                        key: 'cn-huhehaote-ee-vpc',
                        value: 'https://instanceName-registry-vpc.cn-huhehaote.cr.aliyuncs.com',
                    },
                    {
                        key: 'cn-shenzhen-ee-vpc',
                        value: 'https://instanceName-registry-vpc.cn-shenzhen.cr.aliyuncs.com',
                    },
                    {
                        key: 'cn-chengdu-ee-vpc',
                        value: 'https://instanceName-registry-vpc.cn-chengdu.cr.aliyuncs.com',
                    },
                    {
                        key: 'cn-hongkong-ee-vpc',
                        value: 'https://instanceName-registry-vpc.cn-hongkong.cr.aliyuncs.com',
                    },
                    {
                        key: 'ap-southeast-1-ee-vpc',
                        value: 'https://instanceName-registry-vpc.ap-southeast-1.cr.aliyuncs.com',
                    },
                    {
                        key: 'ap-southeast-2-ee-vpc',
                        value: 'https://instanceName-registry-vpc.ap-southeast-2.cr.aliyuncs.com',
                    },
                    {
                        key: 'ap-southeast-3-ee-vpc',
                        value: 'https://instanceName-registry-vpc.ap-southeast-3.cr.aliyuncs.com',
                    },
                    {
                        key: 'ap-southeast-5-ee-vpc',
                        value: 'https://instanceName-registry-vpc.ap-southeast-5.aliyuncs.cr.com',
                    },
                    {
                        key: 'ap-northeast-1-ee-vpc',
                        value: 'https://instanceName-registry-vpc.ap-northeast-1.cr.aliyuncs.com',
                    },
                    {
                        key: 'ap-south-1-ee-vpc',
                        value: 'https:/instanceName-/registry-vpc.ap-south-1.cr.aliyuncs.com',
                    },
                    {
                        key: 'eu-central-1-ee-vpc',
                        value: 'https://instanceName-registry-vpc.eu-central-1.cr.aliyuncs.com',
                    },
                    {
                        key: 'eu-west-1-ee-vpc',
                        value: 'https://instanceName-registry-vpc.eu-west-1.cr.aliyuncs.com',
                    },
                    {
                        key: 'us-west-1-ee-vpc',
                        value: 'https://instanceName-registry-vpc.us-west-1.cr.aliyuncs.com',
                    },
                    {
                        key: 'us-east-1-ee-vpc',
                        value: 'https://instanceName-registry-vpc.us-east-1.cr.aliyuncs.com',
                    },
                    {
                        key: 'me-east-1-ee-vpc',
                        value: 'https://instanceName-registry-vpc.me-east-1.cr.aliyuncs.com',
                    },
                    {
                        key: 'cn-hangzhou-ee',
                        value: `https://instanceName-registry.cn-hangzhou.cr.aliyuncs.com`,
                    },
                    {
                        key: 'cn-shanghai-ee',
                        value: 'https://instanceName-registry.cn-shanghai.cr.aliyuncs.com',
                    },
                    {
                        key: 'cn-qingdao-ee',
                        value: 'https://instanceName-registry.cn-qingdao.cr.aliyuncs.com',
                    },
                    {
                        key: 'cn-beijing-ee',
                        value: 'https://instanceName-registry.cn-beijing.cr.aliyuncs.com',
                    },
                    {
                        key: 'cn-zhangjiakou-ee',
                        value: 'https://instanceName-registry.cn-zhangjiakou.cr.aliyuncs.com',
                    },
                    {
                        key: 'cn-huhehaote-ee',
                        value: 'https://instanceName-registry.cn-huhehaote.cr.aliyuncs.com',
                    },
                    {
                        key: 'cn-shenzhen-ee',
                        value: 'https://instanceName-registry.cn-shenzhen.cr.aliyuncs.com',
                    },
                    {
                        key: 'cn-chengdu-ee',
                        value: 'https://instanceName-registry.cn-chengdu.cr.aliyuncs.com',
                    },
                    {
                        key: 'cn-hongkong-ee',
                        value: 'https://instanceName-registry.cn-hongkong.cr.aliyuncs.com',
                    },
                    {
                        key: 'ap-southeast-1-ee',
                        value: 'https://instanceName-registry.ap-southeast-1.cr.aliyuncs.com',
                    },
                    {
                        key: 'ap-southeast-2-ee',
                        value: 'https://instanceName-registry.ap-southeast-2.cr.aliyuncs.com',
                    },
                    {
                        key: 'ap-southeast-3-ee',
                        value: 'https://instanceName-registry.ap-southeast-3.cr.aliyuncs.com',
                    },
                    {
                        key: 'ap-southeast-5-ee',
                        value: 'https://instanceName-registry.ap-southeast-5.aliyuncs.cr.com',
                    },
                    {
                        key: 'ap-northeast-1-ee',
                        value: 'https://instanceName-registry.ap-northeast-1.cr.aliyuncs.com',
                    },
                    {
                        key: 'ap-south-1-ee',
                        value: 'https:/instanceName-/registry.ap-south-1.cr.aliyuncs.com',
                    },
                    {
                        key: 'eu-central-1-ee',
                        value: 'https://instanceName-registry.eu-central-1.cr.aliyuncs.com',
                    },
                    {
                        key: 'eu-west-1-ee',
                        value: 'https://instanceName-registry.eu-west-1.cr.aliyuncs.com',
                    },
                    {
                        key: 'us-west-1-ee',
                        value: 'https://instanceName-registry.us-west-1.cr.aliyuncs.com',
                    },
                    {
                        key: 'us-east-1-ee',
                        value: 'https://instanceName-registry-.us-east-1.cr.aliyuncs.com',
                    },
                    {
                        key: 'me-east-1-ee',
                        value: 'https://instanceName-registry.me-east-1.cr.aliyuncs.com',
                    },
                ],
            },
            credential_pattern: null,
        },
        'aws-ecr': {
            endpoint_pattern: {
                endpoint_type: 'EndpointPatternTypeList',
                endpoints: [
                    {
                        key: 'ap-northeast-1',
                        value: 'https://api.ecr.ap-northeast-1.amazonaws.com',
                    },
                    {
                        key: 'us-east-1',
                        value: 'https://api.ecr.us-east-1.amazonaws.com',
                    },
                    {
                        key: 'us-east-2',
                        value: 'https://api.ecr.us-east-2.amazonaws.com',
                    },
                    {
                        key: 'us-west-1',
                        value: 'https://api.ecr.us-west-1.amazonaws.com',
                    },
                    {
                        key: 'us-west-2',
                        value: 'https://api.ecr.us-west-2.amazonaws.com',
                    },
                    {
                        key: 'ap-east-1',
                        value: 'https://api.ecr.ap-east-1.amazonaws.com',
                    },
                    {
                        key: 'ap-south-1',
                        value: 'https://api.ecr.ap-south-1.amazonaws.com',
                    },
                    {
                        key: 'ap-northeast-2',
                        value: 'https://api.ecr.ap-northeast-2.amazonaws.com',
                    },
                    {
                        key: 'ap-southeast-1',
                        value: 'https://api.ecr.ap-southeast-1.amazonaws.com',
                    },
                    {
                        key: 'ap-southeast-2',
                        value: 'https://api.ecr.ap-southeast-2.amazonaws.com',
                    },
                    {
                        key: 'ca-central-1',
                        value: 'https://api.ecr.ca-central-1.amazonaws.com',
                    },
                    {
                        key: 'eu-central-1',
                        value: 'https://api.ecr.eu-central-1.amazonaws.com',
                    },
                    {
                        key: 'eu-west-1',
                        value: 'https://api.ecr.eu-west-1.amazonaws.com',
                    },
                    {
                        key: 'eu-west-2',
                        value: 'https://api.ecr.eu-west-2.amazonaws.com',
                    },
                    {
                        key: 'eu-west-3',
                        value: 'https://api.ecr.eu-west-3.amazonaws.com',
                    },
                    {
                        key: 'eu-north-1',
                        value: 'https://api.ecr.eu-north-1.amazonaws.com',
                    },
                    {
                        key: 'sa-east-1',
                        value: 'https://api.ecr.sa-east-1.amazonaws.com',
                    },
                    {
                        key: 'cn-north-1',
                        value: 'https://api.ecr.cn-north-1.amazonaws.com.cn',
                    },
                    {
                        key: 'cn-northwest-1',
                        value: 'https://api.ecr.cn-northwest-1.amazonaws.com.cn',
                    },
                ],
            },
            credential_pattern: null,
        },
        'docker-hub': {
            endpoint_pattern: {
                endpoint_type: 'EndpointPatternTypeFix',
                endpoints: [
                    {
                        key: 'hub.docker.com',
                        value: 'https://hub.docker.com',
                    },
                ],
            },
            credential_pattern: null,
        },
        'google-gcr': {
            endpoint_pattern: {
                endpoint_type: 'EndpointPatternTypeList',
                endpoints: [
                    {
                        key: 'gcr.io',
                        value: 'https://gcr.io',
                    },
                    {
                        key: 'us.gcr.io',
                        value: 'https://us.gcr.io',
                    },
                    {
                        key: 'eu.gcr.io',
                        value: 'https://eu.gcr.io',
                    },
                    {
                        key: 'asia.gcr.io',
                        value: 'https://asia.gcr.io',
                    },
                ],
            },
            credential_pattern: {
                access_key_type: 'AccessKeyTypeFix',
                access_key_data: '_json_key',
                access_secret_type: 'AccessSecretTypeFile',
                access_secret_data: 'No Change',
            },
        },
    };
    let fakedHttp = {
        get() {
            return of(adapterInfoMockData).pipe(delay(0));
        },
    };
    let mockData: Registry[] = [
        {
            id: 1,
            credential: {
                access_key: 'admin',
                access_secret: '',
                type: 'basic',
            },
            description: 'test',
            insecure: false,
            name: 'target_01',
            type: 'Harbor',
            url: 'https://10.117.4.151',
        },
        {
            id: 2,
            credential: {
                access_key: 'AAA',
                access_secret: '',
                type: 'basic',
            },
            description: 'test',
            insecure: false,
            name: 'target_02',
            type: 'Harbor',
            url: 'https://10.117.5.142',
        },
        {
            id: 3,
            credential: {
                access_key: 'admin',
                access_secret: '',
                type: 'basic',
            },
            description: 'test',
            insecure: false,
            name: 'target_03',
            type: 'Harbor',
            url: 'https://101.1.11.111',
        },
        {
            id: 4,
            credential: {
                access_key: 'admin',
                access_secret: '',
                type: 'basic',
            },
            description: 'test',
            insecure: false,
            name: 'target_04',
            type: 'Harbor',
            url: 'https://4.4.4.4',
        },
    ];
    let mockAdapters = ['harbor', 'docker hub'];
    let comp: EndpointComponent;
    let fixture: ComponentFixture<EndpointComponent>;
    const mockAppConfigService = {
        getConfig: () => {
            return {
                project_creation_restriction: '',
                with_chartmuseum: '',
            };
        },
    };
    const mockedEndpointService = {
        getEndpoints(targetName: string) {
            if (targetName) {
                const endpoints: Registry[] = [];
                mockData.forEach(item => {
                    if (item.name.indexOf(targetName) !== -1) {
                        endpoints.push(item);
                    }
                });
                return of(endpoints).pipe(delay(0));
            }
            return of(mockData).pipe(delay(0));
        },
        getAdapters() {
            return of(mockAdapters).pipe(delay(0));
        },
        getEndpointWithReplicationRules() {
            return of([]).pipe(delay(0));
        },
        getEndpoint(endPointId: number | string) {
            if (endPointId) {
                let endpoint: Registry;
                mockData.forEach(item => {
                    if (item.id === endPointId) {
                        endpoint = item;
                    }
                });
                return of(endpoint).pipe(delay(0));
            }
            return of(mockData[0]).pipe(delay(0));
        },
        getAdapterText(adapter: string): string {
            if (ADAPTERS_MAP && ADAPTERS_MAP[adapter]) {
                return ADAPTERS_MAP[adapter];
            }
            return adapter;
        },
    };
    const mockRegistryService = {
        listRegistriesResponse(param?: RegistryService.ListRegistriesParams) {
            if (param && param.q) {
                const endpoints: Registry[] = [];
                mockData.forEach(item => {
                    if (param.q.indexOf(item.name) !== -1) {
                        endpoints.push(item);
                    }
                });
                const response: HttpResponse<Array<Registry>> =
                    new HttpResponse<Array<Registry>>({
                        headers: new HttpHeaders({
                            'x-total-count': endpoints.length.toString(),
                        }),
                        body: endpoints,
                    });
                return of(response).pipe(delay(0));
            }
            const res: HttpResponse<Array<Registry>> = new HttpResponse<
                Array<Registry>
            >({
                headers: new HttpHeaders({ 'x-total-count': '3' }),
                body: mockData,
            });
            return of(res).pipe(delay(0));
        },
    };
    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [CreateEditEndpointComponent, EndpointComponent],
            providers: [
                ErrorHandler,
                { provide: EndpointService, useValue: mockedEndpointService },
                { provide: OperationService },
                { provide: HttpClient, useValue: fakedHttp },
                { provide: AppConfigService, useValue: mockAppConfigService },
                { provide: RegistryService, useValue: mockRegistryService },
            ],
            schemas: [NO_ERRORS_SCHEMA],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(EndpointComponent);
        comp = fixture.componentInstance;
        fixture.autoDetectChanges(true);
    });

    it('should retrieve endpoint data', async () => {
        await fixture.whenStable();
        const rows = fixture.nativeElement.querySelectorAll('clr-dg-row');
        expect(rows.length).toEqual(4);
    });
    it('should open edit endpoint modal', async () => {
        await fixture.whenStable();
        const editButton: HTMLButtonElement =
            fixture.nativeElement.querySelector('#edit');
        comp.selectedRow = [mockData[0]];
        await fixture.whenStable();
        expect(editButton).toBeTruthy();
        editButton.click();
        editButton.dispatchEvent(new Event('click'));
        await fixture.whenStable();
        const nameInput: HTMLInputElement =
            fixture.nativeElement.querySelector('#destination_name');
        expect(nameInput.value).toEqual('target_01');
    });

    it('should filter endpoints by keyword', async () => {
        await fixture.whenStable();
        comp.doSearchTargets('target_02');
        await fixture.whenStable();
        const editButton: HTMLButtonElement =
            fixture.nativeElement.querySelector('#edit');
        comp.selectedRow = [mockData[0]];
        await fixture.whenStable();
        editButton.click();
        editButton.dispatchEvent(new Event('click'));
        await fixture.whenStable();
        expect(comp.targets.length).toEqual(1);
        expect(comp.targets[0].name).toEqual('target_02');
    });
    it('should open creation endpoint', async () => {
        await fixture.whenStable();
        const addButton: HTMLButtonElement =
            fixture.nativeElement.querySelector('#add');
        expect(addButton).toBeTruthy();
        addButton.click();
        addButton.dispatchEvent(new Event('click'));
        await fixture.whenStable();
        const nameInput: HTMLInputElement =
            fixture.nativeElement.querySelector('#destination_name');
        expect(nameInput).toBeTruthy();
    });
});
