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
export class StringValueItem {
    value: string;
    editable: boolean;

    public constructor(v: string, e: boolean) {
        this.value = v;
        this.editable = e;
    }
}

export class NumberValueItem {
    value: number;
    editable: boolean;

    public constructor(v: number, e: boolean) {
        this.value = v;
        this.editable = e;
    }
}

export class BoolValueItem {
    value: boolean;
    editable: boolean;

    public constructor(v: boolean, e: boolean) {
        this.value = v;
        this.editable = e;
    }
}

export class ComplexValueItem {
    value: any | { [key: string]: any | any[] };
    editable: boolean;

    public constructor(v: any | { [key: string]: any | any[] }, e: boolean) {
        this.value = v;
        this.editable = e;
    }
}

export class Configuration {
    [key: string]: any | any[];
    auth_mode: StringValueItem;
    primary_auth_mode: BoolValueItem;
    project_creation_restriction: StringValueItem;
    self_registration: BoolValueItem;
    ldap_base_dn: StringValueItem;
    ldap_filter?: StringValueItem;
    ldap_scope: NumberValueItem;
    ldap_search_dn?: StringValueItem;
    ldap_search_password?: StringValueItem;
    ldap_timeout: NumberValueItem;
    ldap_uid: StringValueItem;
    ldap_url: StringValueItem;
    ldap_verify_cert: BoolValueItem;
    ldap_group_base_dn: StringValueItem;
    ldap_group_search_filter: StringValueItem;
    ldap_group_attribute_name: StringValueItem;
    ldap_group_search_scope: NumberValueItem;
    ldap_group_membership_attribute: StringValueItem;
    ldap_group_admin_dn: StringValueItem;
    uaa_client_id: StringValueItem;
    uaa_client_secret?: StringValueItem;
    uaa_endpoint: StringValueItem;
    uaa_verify_cert: BoolValueItem;
    email_host: StringValueItem;
    email_identity: StringValueItem;
    email_from: StringValueItem;
    email_port: NumberValueItem;
    email_ssl: BoolValueItem;
    email_username?: StringValueItem;
    email_password?: StringValueItem;
    email_insecure: BoolValueItem;
    verify_remote_cert: BoolValueItem;
    robot_token_duration: NumberValueItem;
    token_expiration: NumberValueItem;
    robot_name_prefix?: StringValueItem;
    scan_all_policy: ComplexValueItem;
    read_only: BoolValueItem;
    notification_enable: BoolValueItem;
    http_authproxy_admin_groups?: StringValueItem;
    http_authproxy_endpoint?: StringValueItem;
    http_authproxy_tokenreview_endpoint?: StringValueItem;
    http_authproxy_verify_cert?: BoolValueItem;
    http_authproxy_skip_search?: BoolValueItem;
    oidc_name?: StringValueItem;
    oidc_endpoint?: StringValueItem;
    oidc_client_id?: StringValueItem;
    oidc_client_secret?: StringValueItem;
    oidc_verify_cert?: BoolValueItem;
    oidc_auto_onboard?: BoolValueItem;
    oidc_scope?: StringValueItem;
    oidc_user_claim?: StringValueItem;
    count_per_project: NumberValueItem;
    storage_per_project: NumberValueItem;
    cfg_expiration: NumberValueItem;
    oidc_groups_claim: StringValueItem;
    oidc_admin_group: StringValueItem;
    oidc_group_filter: StringValueItem;
    audit_log_forward_endpoint: StringValueItem;
    skip_audit_log_database: BoolValueItem;
    audit_log_track_ip_address: BoolValueItem;
    audit_log_track_user_agent: BoolValueItem;
    session_timeout: NumberValueItem;
    scanner_skip_update_pulltime: BoolValueItem;
    banner_message: StringValueItem;
    public constructor() {
        this.auth_mode = new StringValueItem('db_auth', true);
        this.primary_auth_mode = new BoolValueItem(false, true);
        this.project_creation_restriction = new StringValueItem(
            'everyone',
            true
        );
        this.self_registration = new BoolValueItem(false, true);
        this.ldap_base_dn = new StringValueItem('', true);
        this.ldap_filter = new StringValueItem('', true);
        this.ldap_scope = new NumberValueItem(0, true);
        this.ldap_search_dn = new StringValueItem('', true);
        this.ldap_search_password = new StringValueItem('', true);
        this.ldap_timeout = new NumberValueItem(5, true);
        this.ldap_uid = new StringValueItem('', true);
        this.ldap_url = new StringValueItem('', true);
        this.ldap_verify_cert = new BoolValueItem(true, true);
        this.ldap_group_base_dn = new StringValueItem('', true);
        this.ldap_group_search_filter = new StringValueItem('', true);
        this.ldap_group_attribute_name = new StringValueItem('', true);
        this.ldap_group_search_scope = new NumberValueItem(0, true);
        this.ldap_group_membership_attribute = new StringValueItem('', true);
        this.ldap_group_admin_dn = new StringValueItem('', true);
        this.uaa_client_id = new StringValueItem('', true);
        this.uaa_client_secret = new StringValueItem('', true);
        this.uaa_endpoint = new StringValueItem('', true);
        this.uaa_verify_cert = new BoolValueItem(false, true);
        this.email_host = new StringValueItem('', true);
        this.email_identity = new StringValueItem('', true);
        this.email_from = new StringValueItem('', true);
        this.email_port = new NumberValueItem(25, true);
        this.email_ssl = new BoolValueItem(false, true);
        this.email_username = new StringValueItem('', true);
        this.email_password = new StringValueItem('', true);
        this.email_insecure = new BoolValueItem(false, true);
        this.token_expiration = new NumberValueItem(30, true);
        this.robot_name_prefix = new StringValueItem('', true);
        this.robot_token_duration = new NumberValueItem(30, true);
        this.cfg_expiration = new NumberValueItem(30, true);
        this.verify_remote_cert = new BoolValueItem(false, true);
        this.scan_all_policy = new ComplexValueItem(
            {
                type: 'daily',
                parameter: {
                    daily_time: 0,
                },
            },
            true
        );
        this.read_only = new BoolValueItem(false, true);
        this.notification_enable = new BoolValueItem(false, true);
        this.http_authproxy_admin_groups = new StringValueItem('', true);
        this.http_authproxy_endpoint = new StringValueItem('', true);
        this.http_authproxy_tokenreview_endpoint = new StringValueItem(
            '',
            true
        );
        this.http_authproxy_verify_cert = new BoolValueItem(false, true);
        this.http_authproxy_skip_search = new BoolValueItem(false, true);
        this.oidc_name = new StringValueItem('', true);
        this.oidc_endpoint = new StringValueItem('', true);
        this.oidc_client_id = new StringValueItem('', true);
        this.oidc_client_secret = new StringValueItem('', true);
        this.oidc_verify_cert = new BoolValueItem(false, true);
        this.oidc_auto_onboard = new BoolValueItem(false, true);
        this.oidc_scope = new StringValueItem('', true);
        this.oidc_groups_claim = new StringValueItem('', true);
        this.oidc_admin_group = new StringValueItem('', true);
        this.oidc_group_filter = new StringValueItem('', true);
        this.oidc_user_claim = new StringValueItem('', true);
        this.count_per_project = new NumberValueItem(-1, true);
        this.storage_per_project = new NumberValueItem(-1, true);
        this.audit_log_forward_endpoint = new StringValueItem('', true);
        this.skip_audit_log_database = new BoolValueItem(false, true);
        this.audit_log_track_ip_address = new BoolValueItem(false, true);
        this.audit_log_track_user_agent = new BoolValueItem(false, true);
        this.session_timeout = new NumberValueItem(60, true);
        this.scanner_skip_update_pulltime = new BoolValueItem(false, true);
        this.banner_message = new StringValueItem(
            JSON.stringify(new BannerMessage()),
            true
        );
    }
}

export class ScanningMetrics {
    total?: number;
    completed?: number;
    metrics: {
        [key: string]: number;
    };
    requester?: string;
    trigger?: string;
    ongoing: boolean;
}
export enum Triggers {
    MANUAL = 'Manual',
    SCHEDULE = 'Schedule',
    EVENT = 'Event',
}

export class BannerMessage {
    message: string;
    closable: boolean;
    type: string;
    fromDate: Date;
    toDate: Date;
    constructor() {
        this.closable = false;
    }
}

export enum BannerMessageType {
    SUCCESS = 'success',
    INFO = 'info',
    WARNING = 'warning',
    ERROR = 'danger',
}

export const BannerMessageI18nMap = {
    [BannerMessageType.SUCCESS]: 'BANNER_MESSAGE.SUCCESS',
    [BannerMessageType.INFO]: 'BANNER_MESSAGE.INFO',
    [BannerMessageType.WARNING]: 'BANNER_MESSAGE.WARNING',
    [BannerMessageType.ERROR]: 'BANNER_MESSAGE.DANGER',
};
