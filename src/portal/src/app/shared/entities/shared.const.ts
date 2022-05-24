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

export const enum AlertType {
    DANGER,
    WARNING,
    INFO,
    SUCCESS,
}

export const dismissInterval = 10 * 1000;
export const httpStatusCode = {
    Unauthorized: 401,
    Forbidden: 403,
};
export const enum ConfirmationTargets {
    EMPTY,
    PROJECT,
    PROJECT_MEMBER,
    USER,
    ROBOT_ACCOUNT,
    POLICY,
    TOGGLE_CONFIRM,
    TARGET,
    REPOSITORY,
    TAG,
    CONFIG,
    CONFIG_ROUTE,
    CONFIG_TAB,
    HELM_CHART,
    HELM_CHART_VERSION,
    STOP_EXECUTIONS,
    SCANNER,
    REPLICATION,
    ROBOT_ACCOUNT_ENABLE_OR_DISABLE,
    INSTANCE,
    P2P_PROVIDER,
    P2P_PROVIDER_STOP,
    P2P_PROVIDER_EXECUTE,
    P2P_PROVIDER_DELETE,
    PROJECT_ROBOT_ACCOUNT,
    PROJECT_ROBOT_ACCOUNT_ENABLE_OR_DISABLE,
    WEBHOOK,
    ACCESSORY,
    ALL_ACCESSORIES,
}

export const enum ActionType {
    ADD_NEW,
    EDIT,
}

export const ListMode = {
    READONLY: 'readonly',
    FULL: 'full',
};

export const CommonRoutes = {
    SIGN_IN: '/sign-in',
    EMBEDDED_SIGN_IN: '/account/sign-in',
    SIGN_UP: '/sign-in?sign_up=true',
    EMBEDDED_SIGN_UP: '/account/sign-in?sign_up=true',
    HARBOR_ROOT: '/harbor',
    HARBOR_DEFAULT: '/harbor/projects',
};

export const enum ConfirmationState {
    NA,
    CONFIRMED,
    CANCEL,
}
export const FilterType = {
    NAME: 'name',
    TAG: 'tag',
    LABEL: 'label',
    RESOURCE: 'resource',
};

export const enum ConfirmationButtons {
    CONFIRM_CANCEL,
    YES_NO,
    DELETE_CANCEL,
    CLOSE,
    ENABLE_CANCEL,
    DISABLE_CANCEL,
    REPLICATE_CANCEL,
    STOP_CANCEL,
}
export const QuotaUnits = [
    {
        UNIT: 'Byte',
    },
    {
        UNIT: 'KiB',
    },
    {
        UNIT: 'MiB',
    },
    {
        UNIT: 'GiB',
    },
    {
        UNIT: 'TiB',
    },
];
export const QuotaUnlimited = -1;
export const StorageMultipleConstant = 1024;
export const LimitCount = 100000000;
export enum QuotaUnit {
    TB = 'TiB',
    GB = 'GiB',
    MB = 'MiB',
    KB = 'KiB',
    BIT = 'Byte',
}
export enum QuotaProgress {
    COUNT_USED = 'COUNT_USED',
    COUNT_HARD = 'COUNT_HARD',
    STROAGE_USED = 'STORAGE_USED',
    STORAGE_HARD = 'STORAGE_HARD',
}

export const LabelColor = [
    { color: '#000000', textColor: 'white' },
    { color: '#61717D', textColor: 'white' },
    { color: '#737373', textColor: 'white' },
    { color: '#80746D', textColor: 'white' },
    { color: '#FFFFFF', textColor: 'black' },
    { color: '#A9B6BE', textColor: 'black' },
    { color: '#DDDDDD', textColor: 'black' },
    { color: '#BBB3A9', textColor: 'black' },
    { color: '#0065AB', textColor: 'white' },
    { color: '#343DAC', textColor: 'white' },
    { color: '#781DA0', textColor: 'white' },
    { color: '#9B0D54', textColor: 'white' },
    { color: '#0095D3', textColor: 'black' },
    { color: '#9DA3DB', textColor: 'black' },
    { color: '#BE90D6', textColor: 'black' },
    { color: '#F1428A', textColor: 'black' },
    { color: '#1D5100', textColor: 'white' },
    { color: '#006668', textColor: 'white' },
    { color: '#006690', textColor: 'white' },
    { color: '#004A70', textColor: 'white' },
    { color: '#48960C', textColor: 'black' },
    { color: '#00AB9A', textColor: 'black' },
    { color: '#00B7D6', textColor: 'black' },
    { color: '#0081A7', textColor: 'black' },
    { color: '#C92100', textColor: 'white' },
    { color: '#CD3517', textColor: 'white' },
    { color: '#C25400', textColor: 'white' },
    { color: '#D28F00', textColor: 'white' },
    { color: '#F52F52', textColor: 'black' },
    { color: '#FF5501', textColor: 'black' },
    { color: '#F57600', textColor: 'black' },
    { color: '#FFDC0B', textColor: 'black' },
];

export const CONFIG_AUTH_MODE = {
    HTTP_AUTH: 'http_auth',
    LDAP_AUTH: 'ldap_auth',
    OIDC_AUTH: 'oidc_auth',
    UAA_AUTH: 'uaa_auth',
    DB_AUTH: 'db_auth',
};
export const QUOTA_DANGER_COEFFICIENT = 0.9;
export const QUOTA_WARNING_COEFFICIENT = 0.7;
export const PROJECT_ROOTS = [
    {
        NAME: 'admin',
        VALUE: 1,
        LABEL: 'GROUP.PROJECT_ADMIN',
    },
    {
        NAME: 'maintainer',
        VALUE: 4,
        LABEL: 'GROUP.PROJECT_MAINTAINER',
    },
    {
        NAME: 'developer',
        VALUE: 2,
        LABEL: 'GROUP.DEVELOPER',
    },
    {
        NAME: 'guest',
        VALUE: 3,
        LABEL: 'GROUP.GUEST',
    },
    {
        NAME: 'limited',
        VALUE: 5,
        LABEL: 'GROUP.LIMITED_GUEST',
    },
];

export enum GroupType {
    LDAP_TYPE = 1,
    HTTP_TYPE = 2,
    OIDC_TYPE = 3,
}
export const REFRESH_TIME_DIFFERENCE = 10000;

//

export const DeFaultLang = 'en-us';
export type SupportedLanguage = keyof typeof LANGUAGES;
export const LANGUAGES = {
    'en-us': 'English',
    'zh-cn': '中文简体',
    'zh-tw': '中文繁體',
    'es-es': 'Español',
    'fr-fr': 'Français',
    'pt-br': 'Português do Brasil',
    'tr-tr': 'Türkçe',
    'de-de': 'Deutsch',
} as const;
export const supportedLangs = Object.keys(LANGUAGES) as SupportedLanguage[];
/**
 * The default cookie key used to store current used language preference.
 */
export const DEFAULT_LANG_LOCALSTORAGE_KEY = 'harbor-lang';

export type DatetimeRendering = keyof typeof DATETIME_RENDERINGS;
export const DATETIME_RENDERINGS = {
    'locale-default': 'TOP_NAV.DATETIME_RENDERING_DEFAULT',
    'iso-8601': 'ISO 8601',
} as const;
export const DefaultDatetimeRendering = 'locale-default';
/**
 * The default cookie key used to store current used datetime rendering preference.
 */
export const DEFAULT_DATETIME_RENDERING_LOCALSTORAGE_KEY =
    'harbor-datetime-rendering';

export const AdmiralQueryParamKey = 'admiral_redirect_url';

export const HarborQueryParamKey = 'harbor_redirect_url';

export const CookieKeyOfAdmiral = 'admiral.endpoint.latest';

export const ProjectTypes = {
    0: 'PROJECT.ALL_PROJECTS',
    1: 'PROJECT.PRIVATE_PROJECTS',
    2: 'PROJECT.PUBLIC_PROJECTS',
};

export const RoleInfo = {
    1: 'MEMBER.PROJECT_ADMIN',
    2: 'MEMBER.DEVELOPER',
    3: 'MEMBER.GUEST',
    4: 'MEMBER.PROJECT_MAINTAINER',
    5: 'MEMBER.LIMITED_GUEST',
};

export const RoleMapping = {
    projectAdmin: 'MEMBER.PROJECT_ADMIN',
    maintainer: 'MEMBER.PROJECT_MAINTAINER',
    developer: 'MEMBER.DEVELOPER',
    guest: 'MEMBER.GUEST',
    limitedGuest: 'MEMBER.LIMITED_GUEST',
};

export const ProjectRoles = [
    { id: 1, value: 'MEMBER.PROJECT_ADMIN' },
    { id: 2, value: 'MEMBER.DEVELOPER' },
    { id: 3, value: 'MEMBER.GUEST' },
    { id: 4, value: 'MEMBER.PROJECT_MAINTAINER' },
    { id: 5, value: 'MEMBER.LIMITED_GUEST' },
];

export enum Roles {
    PROJECT_ADMIN = 1,
    PROJECT_MAINTAINER = 4,
    DEVELOPER = 2,
    GUEST = 3,
    LIMITED_GUEST = 5,
    OTHER = 0,
}
export const DefaultHelmIcon = '/images/helm-gray.svg';

export enum ResourceType {
    REPOSITORY = 1,
    CHART_VERSION = 2,
    REPOSITORY_TAG = 3,
}

export const TRUE_STR: string = 'true';
export const FALSE_STR: string = 'false';

export const CARD_VIEW_LOCALSTORAGE_KEY = 'card-view';

export const PROJECT_SUMMARY_CARD_VIEW_LOCALSTORAGE_KEY = 'project_card-view';

export enum ScheduleType {
    NONE = 'None',
    DAILY = 'Daily',
    WEEKLY = 'Weekly',
    HOURLY = 'Hourly',
    CUSTOM = 'Custom',
    MANUAL = 'Manual',
}
