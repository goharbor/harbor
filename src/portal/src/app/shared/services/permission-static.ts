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
export const USERSTATICPERMISSION = {
    PROJECT: {
        KEY: '.',
        VALUE: {
            DELETE: 'delete',
            UPDATE: 'update',
            READ: 'read',
        },
    },
    MEMBER: {
        KEY: 'member',
        VALUE: {
            CREATE: 'create',
            UPDATE: 'update',
            DELETE: 'delete',
            READ: 'read',
            LIST: 'list',
        },
    },
    LOG: {
        KEY: 'log',
        VALUE: {
            LIST: 'list',
        },
    },
    // to do remove
    REPLICATION: {
        KEY: 'replication',
        VALUE: {
            CREATE: 'create',
            UPDATE: 'update',
            DELETE: 'delete',
            LIST: 'list',
        },
    },
    // to do remove
    REPLICATION_JOB: {
        KEY: 'replication-job',
        VALUE: {
            CREATE: 'create',
        },
    },
    LABEL: {
        KEY: 'label',
        VALUE: {
            CREATE: 'create',
            UPDATE: 'update',
            DELETE: 'delete',
            READ: 'read',
            LIST: 'list',
        },
    },
    CONFIGURATION: {
        KEY: 'configuration',
        VALUE: {
            UPDATE: 'update',
            READ: 'read',
        },
    },
    QUOTA: {
        KEY: 'quota',
        VALUE: {
            READ: 'read',
        },
    },
    REPOSITORY: {
        KEY: 'repository',
        VALUE: {
            CREATE: 'create',
            UPDATE: 'update',
            DELETE: 'delete',
            LIST: 'list',
            PUSH: 'push',
            READ: 'read',
            PULL: 'pull',
        },
    },
    ARTIFACT: {
        KEY: 'artifact',
        VALUE: {
            CREATE: 'create',
            DELETE: 'delete',
            LIST: 'list',
            READ: 'read',
        },
    },
    ARTIFACT_ADDITION: {
        KEY: 'artifact-addition',
        VALUE: {
            READ: 'read',
        },
    },
    REPOSITORY_TAG: {
        KEY: 'tag',
        VALUE: {
            DELETE: 'delete',
            LIST: 'list',
            CREATE: 'create',
        },
    },
    REPOSITORY_TAG_SCAN_JOB: {
        KEY: 'scan',
        VALUE: {
            CREATE: 'create',
            READ: 'read',
        },
    },
    REPOSITORY_TAG_SBOM_JOB: {
        KEY: 'sbom',
        VALUE: {
            CREATE: 'create',
            READ: 'read',
        },
    },
    REPOSITORY_ARTIFACT_LABEL: {
        KEY: 'artifact-label',
        VALUE: {
            CREATE: 'create',
            DELETE: 'delete',
        },
    },
    ROBOT: {
        KEY: 'robot',
        VALUE: {
            CREATE: 'create',
            UPDATE: 'update',
            DELETE: 'delete',
            LIST: 'list',
            READ: 'read',
        },
    },
    TAG_RETENTION: {
        KEY: 'tag-retention',
        VALUE: {
            CREATE: 'create',
            UPDATE: 'update',
            DELETE: 'delete',
            LIST: 'list',
            READ: 'read',
            OPERATE: 'operate',
        },
    },
    IMMUTABLE_TAG: {
        KEY: 'immutable-tag',
        VALUE: {
            CREATE: 'create',
            UPDATE: 'update',
            DELETE: 'delete',
            LIST: 'list',
        },
    },
    WEBHOOK: {
        KEY: 'notification-policy',
        VALUE: {
            LIST: 'list',
            READ: 'read',
            CREATE: 'create',
            UPDATE: 'update',
        },
    },
    P2P_PROVIDER: {
        KEY: 'preheat-policy',
        VALUE: {
            LIST: 'list',
            READ: 'read',
            CREATE: 'create',
            UPDATE: 'update',
            DELETE: 'delete',
        },
    },
    SCANNER: {
        KEY: 'scanner',
        VALUE: {
            READ: 'read',
            CREATE: 'create',
        },
    },
    METADATA: {
        KEY: 'metadata',
        VALUE: {
            READ: 'read',
            CREATE: 'create',
            UPDATE: 'update',
            DELETE: 'delete',
        },
    },
    EXPORT_CVE: {
        KEY: 'export-cve',
        VALUE: {
            READ: 'read',
            CREATE: 'create',
            LIST: 'list',
        },
    },
};
