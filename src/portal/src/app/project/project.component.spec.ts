import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { ProjectComponent } from './project.component';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { BrowserAnimationsModule, NoopAnimationsModule } from '@angular/platform-browser/animations';
import { ClarityModule } from '@clr/angular';
import { FormsModule } from '@angular/forms';
import { RouterTestingModule } from '@angular/router/testing';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { ConfigurationService } from '../config/config.service';
import { SessionService } from "../shared/session.service";
import { of } from 'rxjs';
describe('ProjectComponent', () => {
    let component: ProjectComponent;
    let fixture: ComponentFixture<ProjectComponent>;
    const mockSessionService = {
        getCurrentUser: () => {
            return of({
                user_id: 1
            });
        }
    };
    const mockConfigurationService = {
        getConfiguration: () => {
            return of({
                "auth_mode": {
                  "value": "oidc_auth",
                  "editable": false
                },
                "count_per_project": {
                  "value": -1,
                  "editable": true
                },
                "email_from": {
                  "value": "admin \u003csamplin@mydomain.com\u003e",
                  "editable": true
                },
                "email_host": {
                  "value": "smtp.mydomain.com",
                  "editable": true
                },
                "email_identity": {
                  "value": "",
                  "editable": true
                },
                "email_insecure": {
                  "value": false,
                  "editable": true
                },
                "email_port": {
                  "value": 25,
                  "editable": true
                },
                "email_ssl": {
                  "value": false,
                  "editable": true
                },
                "email_username": {
                  "value": "sample_admin@mydomain.com",
                  "editable": true
                },
                "http_authproxy_endpoint": {
                  "value": "",
                  "editable": true
                },
                "http_authproxy_skip_search": {
                  "value": false,
                  "editable": true
                },
                "http_authproxy_tokenreview_endpoint": {
                  "value": "",
                  "editable": true
                },
                "http_authproxy_verify_cert": {
                  "value": true,
                  "editable": true
                },
                "ldap_base_dn": {
                  "value": "",
                  "editable": true
                },
                "ldap_filter": {
                  "value": "",
                  "editable": true
                },
                "ldap_group_admin_dn": {
                  "value": "",
                  "editable": true
                },
                "ldap_group_attribute_name": {
                  "value": "",
                  "editable": true
                },
                "ldap_group_base_dn": {
                  "value": "",
                  "editable": true
                },
                "ldap_group_membership_attribute": {
                  "value": "memberof",
                  "editable": true
                },
                "ldap_group_search_filter": {
                  "value": "",
                  "editable": true
                },
                "ldap_group_search_scope": {
                  "value": 2,
                  "editable": true
                },
                "ldap_scope": {
                  "value": 2,
                  "editable": true
                },
                "ldap_search_dn": {
                  "value": "",
                  "editable": true
                },
                "ldap_timeout": {
                  "value": 5,
                  "editable": true
                },
                "ldap_uid": {
                  "value": "cn",
                  "editable": true
                },
                "ldap_url": {
                  "value": "",
                  "editable": true
                },
                "ldap_verify_cert": {
                  "value": true,
                  "editable": true
                },
                "notification_enable": {
                  "value": true,
                  "editable": true
                },
                "oidc_client_id": {
                  "value": "harb-https",
                  "editable": true
                },
                "oidc_endpoint": {
                  "value": "https://10.158..96:5554/dex",
                  "editable": true
                },
                "oidc_groups_claim": {
                  "value": "",
                  "editable": true
                },
                "oidc_name": {
                  "value": "dex",
                  "editable": true
                },
                "oidc_scope": {
                  "value": "openid,profilline_access",
                  "editable": true
                },
                "oidc_verify_cert": {
                  "value": false,
                  "editable": true
                },
                "project_creation_restriction": {
                  "value": "everyone",
                  "editable": true
                },
                "quota_per_project_enable": {
                  "value": true,
                  "editable": true
                },
                "read_only": {
                  "value": false,
                  "editable": true
                },
                "robot_token_duration": {
                  "value": 43200,
                  "editable": true
                },
                "scan_all_policy": {
                  "value": null,
                  "editable": true
                },
                "self_registration": {
                  "value": false,
                  "editable": true
                },
                "storage_per_project": {
                  "value": -1,
                  "editable": true
                },
                "token_expiration": {
                  "value": 30,
                  "editable": true
                },
                "uaa_client_id": {
                  "value": "",
                  "editable": true
                },
                "uaa_client_secret": {
                  "value": "",
                  "editable": true
                },
                "uaa_endpoint": {
                  "value": "",
                  "editable": true
                },
                "uaa_verify_cert": {
                  "value": false,
                  "editable": true
                }
              });
        }
    };
    beforeEach(async(() => {
        TestBed.configureTestingModule({
            schemas: [
                CUSTOM_ELEMENTS_SCHEMA
            ],
            imports: [
                BrowserAnimationsModule,
                ClarityModule,
                TranslateModule.forRoot(),
                FormsModule,
                RouterTestingModule,
                NoopAnimationsModule,
                HttpClientTestingModule
            ],
            declarations: [ProjectComponent],
            providers: [
                TranslateService,
                { provide: SessionService, useValue: mockSessionService },
                { provide: ConfigurationService, useValue: mockConfigurationService },

            ]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(ProjectComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
