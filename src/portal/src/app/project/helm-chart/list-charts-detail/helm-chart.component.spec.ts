import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { HelmChartComponent } from './helm-chart.component';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { ClarityModule } from '@clr/angular';
import { FormsModule } from '@angular/forms';
import { of } from 'rxjs';
import { HelmChartService } from "../helm-chart.service";
import { ErrorHandler } from "../../../../lib/utils/error-handler";
import { SystemInfoService, UserPermissionService } from "../../../../lib/services";
import { OperationService } from "../../../../lib/components/operation/operation.service";

describe('HelmChartComponent', () => {
    let component: HelmChartComponent;
    let fixture: ComponentFixture<HelmChartComponent>;
    const mockErrorHandler = null;
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
        },
        getHelmCharts() {
            return of([]);
        },
    };
    const mockUserPermissionService = {
        getPermission() {
            return of(true);
        }
    };
    const mockOperationService = {
        publishInfo: () => {
            return of([]);
        },
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
            declarations: [HelmChartComponent],
            providers: [
                TranslateService,
                { provide: ErrorHandler, useValue: mockErrorHandler },
                { provide: SystemInfoService, useValue: mockSystemInfoService },
                { provide: HelmChartService, useValue: mockHelmChartService },
                { provide: UserPermissionService, useValue: mockUserPermissionService },
                { provide: OperationService, useValue: mockOperationService },

            ]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(HelmChartComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
