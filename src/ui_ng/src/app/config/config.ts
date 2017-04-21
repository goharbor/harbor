// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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

export class Configuration {
    auth_mode: StringValueItem;
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
    email_host: StringValueItem;
    email_identity: StringValueItem;
    email_from: StringValueItem;
    email_port: NumberValueItem;
    email_ssl: BoolValueItem;
    email_username?: StringValueItem;
    email_password?: StringValueItem;
    verify_remote_cert: BoolValueItem;
    token_expiration: NumberValueItem;
    cfg_expiration: NumberValueItem;

    public constructor() {
        this.auth_mode = new StringValueItem("db_auth", true);
        this.project_creation_restriction = new StringValueItem("everyone", true);
        this.self_registration = new BoolValueItem(false, true);
        this.ldap_base_dn = new StringValueItem("", true);
        this.ldap_filter = new StringValueItem("", true);
        this.ldap_scope = new NumberValueItem(0, true);
        this.ldap_search_dn = new StringValueItem("", true);
        this.ldap_search_password = new StringValueItem("", true);
        this.ldap_timeout = new NumberValueItem(5, true);
        this.ldap_uid = new StringValueItem("", true);
        this.ldap_url = new StringValueItem("", true);
        this.email_host = new StringValueItem("", true);
        this.email_identity = new StringValueItem("", true);
        this.email_from = new StringValueItem("", true);
        this.email_port = new NumberValueItem(25, true);
        this.email_ssl = new BoolValueItem(false, true);
        this.email_username = new StringValueItem("", true);
        this.email_password = new StringValueItem("", true);
        this.token_expiration = new NumberValueItem(5, true);
        this.cfg_expiration = new NumberValueItem(30, true);
        this.verify_remote_cert = new BoolValueItem(false, true);
    }
}