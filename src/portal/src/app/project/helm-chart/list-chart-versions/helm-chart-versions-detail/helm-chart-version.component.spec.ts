import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { ChartVersionComponent } from './helm-chart-version.component';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { ClarityModule } from '@clr/angular';
import { FormsModule } from '@angular/forms';
import { HelmChartService } from "../../helm-chart.service";
import { LabelFilterComponent } from "../../label-filter/label-filter.component";

import {
    ErrorHandler,
    SystemInfoService,
    LabelService,
    OperationService,
    UserPermissionService
} from "@harbor/ui";
import { of } from 'rxjs';
describe('ChartVersionComponent', () => {
    let component: ChartVersionComponent;
    let fixture: ComponentFixture<ChartVersionComponent>;
    const mockSystemInfoService = {
        getSystemInfo: () => {
            return of(
                {
                    "with_notary": false,
                    "with_admiral": false,
                    "admiral_endpoint": "",
                    "auth_mode": "oidc_auth",
                    "registry_url": "nightly-oidc.harbor.io",
                    "external_url": "https://nightly-oidc.harbor.io",
                    "project_creation_restriction": "everyone",
                    "self_registration": false,
                    "has_ca_root": false,
                    "harbor_version": "dev",
                    "registry_storage_provider_name": "filesystem",
                    "read_only": false,
                    "with_chartmuseum": true,
                    "notification_enable": true
                }
            );
        }
    };
    const mockLabelService = {
        getLabels: () => {
            return of([]);
        },
        getProjectLabels: () => {
            return of([]);
        },
    };
    const mockErrorHandler = null;
    const mockOperationService = {
        publishInfo: () => {
            return of([]);
        },
    };
    const mockUserPermissionService = {
        getPermission() {
            return of(true);
        }
    };
    const mockHelmChartService = {
        getChartVersions() {
            return of(
                [{
                    name: "string",
                    home: "string",
                    sources: [],
                    version: "string",
                    description: "string",
                    keywords: [],
                    maintainers: [],
                    engine: "string",
                    icon: "string",
                    appVersion: "string",
                    apiVersion: "string",
                    urls: [],
                    created: "string",
                    digest: "string",
                    labels: []
                }]
            );
        }
    };
    beforeEach(async(() => {
        TestBed.configureTestingModule({
            schemas: [
                CUSTOM_ELEMENTS_SCHEMA
            ],
            imports: [
                ClarityModule,
                TranslateModule.forRoot(),
                FormsModule
            ],
            declarations: [ChartVersionComponent, LabelFilterComponent],
            providers: [
                TranslateService,
                { provide: SystemInfoService, useValue: mockSystemInfoService },
                { provide: LabelService, useValue: mockLabelService },
                { provide: UserPermissionService, useValue: mockUserPermissionService },
                { provide: ErrorHandler, useValue: mockErrorHandler },
                { provide: HelmChartService, useValue: mockHelmChartService },
                { provide: OperationService, useValue: mockOperationService },
            ]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(ChartVersionComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
