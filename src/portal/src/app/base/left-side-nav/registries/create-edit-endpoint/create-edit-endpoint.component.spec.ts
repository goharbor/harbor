import { ComponentFixture, TestBed } from '@angular/core/testing';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';
import { FilterComponent } from '../../../../shared/components/filter/filter.component';
import { CreateEditEndpointComponent } from './create-edit-endpoint.component';
import { InlineAlertComponent } from '../../../../shared/components/inline-alert/inline-alert.component';
import { ErrorHandler } from '../../../../shared/units/error-handler';
import { Endpoint } from '../../../../shared/services';
import {
    EndpointService,
    EndpointDefaultService,
} from '../../../../shared/services/endpoint.service';
import { of } from 'rxjs';
import { HttpClient } from '@angular/common/http';
import { AppConfigService } from '../../../../services/app-config.service';
import { SharedTestingModule } from '../../../../shared/shared.module';

describe('CreateEditEndpointComponent (inline template)', () => {
    let mockData: Endpoint = {
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
    };
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
            return of(adapterInfoMockData);
        },
    };
    let mockAdapters = ['harbor', 'docker hub'];
    let comp: CreateEditEndpointComponent;
    let fixture: ComponentFixture<CreateEditEndpointComponent>;
    let endpointService: EndpointService;
    let http: HttpClient;
    let spy: jasmine.Spy;
    let spyAdapter: jasmine.Spy;
    const mockAppConfigService = {
        getConfig: () => {
            return {
                project_creation_restriction: '',
                with_chartmuseum: '',
            };
        },
    };
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [SharedTestingModule, NoopAnimationsModule],
            declarations: [
                FilterComponent,
                CreateEditEndpointComponent,
                InlineAlertComponent,
            ],
            providers: [
                ErrorHandler,
                { provide: EndpointService, useClass: EndpointDefaultService },
                { provide: HttpClient, useValue: fakedHttp },
                { provide: AppConfigService, useValue: mockAppConfigService },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(CreateEditEndpointComponent);
        comp = fixture.componentInstance;
        endpointService = fixture.debugElement.injector.get(EndpointService);
        http = fixture.debugElement.injector.get(HttpClient);
        spyAdapter = spyOn(endpointService, 'getAdapters').and.returnValue(
            of(mockAdapters)
        );
        spy = spyOn(endpointService, 'getEndpoint').and.returnValue(
            of(mockData)
        );
        fixture.detectChanges();

        comp.openCreateEditTarget(true, 1);
        fixture.detectChanges();
    });

    it('should be created', () => {
        fixture.detectChanges();
        expect(comp).toBeTruthy();
    });

    it('should get endpoint be called', () => {
        fixture.detectChanges();
        fixture.whenStable().then(() => {
            fixture.detectChanges();
            expect(spy.calls.any()).toBeTruthy();
        });
    });
    it('should get adapterInfo', () => {
        fixture.detectChanges();
        fixture.whenStable().then(() => {
            fixture.detectChanges();
            expect(comp.adapterInfo).toBeTruthy();
        });
    });

    it('should get endpoint and open modal', () => {
        fixture.detectChanges();
        fixture.whenStable().then(() => {
            fixture.detectChanges();
            expect(comp.target.name).toEqual('target_01');
        });
    });
});
