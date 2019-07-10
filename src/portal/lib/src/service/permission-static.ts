export const USERSTATICPERMISSION = {
    "PROJECT": {
        'KEY': '.',
        'VALUE': {
            "DELETE": "delete",
            "UPDATE": "update",
            "READ": "read",
        }
    },
    "MEMBER": {
        'KEY': 'member',
        'VALUE': {
            "CREATE": "create",
            "UPDATE": "update",
            "DELETE": "delete",
            "LIST": "list"
        }
    },
    "LOG": {
        'KEY': 'log',
        'VALUE': {
            "LIST": "list"
        }
    },
    "REPLICATION": {
        'KEY': 'replication',
        'VALUE': {
            "CREATE": "create",
            "UPDATE": "update",
            "DELETE": "delete",
            "LIST": "list",
        }
    },
    "REPLICATION_JOB": {
        'KEY': 'replication-job',
        'VALUE': {
            "CREATE": "create",
        }
    },
    "LABEL": {
        'KEY': 'label',
        'VALUE': {
            "CREATE": "create",
            "UPDATE": "update",
            "DELETE": "delete",
            "LIST": "list",
        }
    },
    "CONFIGURATION": {
        'KEY': 'configuration',
        'VALUE': {
            "UPDATE": "update",
            "READ": "read",
        }
    },
    "REPOSITORY": {
        'KEY': 'repository',
        'VALUE': {
            "CREATE": "create",
            "UPDATE": "update",
            "DELETE": "delete",
            "LIST": "list",
            "PUSH": "push",
            "PULL": "pull",
        }
    },
    "REPOSITORY_TAG": {
        'KEY': 'repository-tag',
        'VALUE': {
            "DELETE": "delete",
            "LIST": "list",
        }
    },
    "REPOSITORY_TAG_SCAN_JOB": {
        'KEY': 'repository-tag-scan-job',
        'VALUE': {
            "CREATE": "create",
            "READ": "read",
            "LIST": "list",
        }
    },
    "REPOSITORY_TAG_VULNERABILITY": {
        'KEY': 'repository-tag-vulnerability',
        'VALUE': {
            "LIST": "list",
        }
    },
    "REPOSITORY_TAG_LABEL": {
        'KEY': 'repository-tag-label',
        'VALUE': {
            "CREATE": "create",
            "DELETE": "delete",
        }
    },
    "REPOSITORY_TAG_MANIFEST": {
        'KEY': 'repository-tag-manifest',
        'VALUE': {
            "READ": "read",
        }
    },
    "HELM_CHART": {
        'KEY': 'helm-chart',
        'VALUE': {
            "UPLOAD": "create",
            "DOWNLOAD": "read",
            "DELETE": "delete",
            "LIST": "list",
        }
    },
    "HELM_CHART_VERSION": {
        'KEY': 'helm-chart-version',
        'VALUE': {
            "DELETE": "delete",
            "LIST": "list",
            "READ": "read",
        }
    },
    "HELM_CHART_VERSION_LABEL": {
        'KEY': 'helm-chart-version-label',
        'VALUE': {
            "CREATE": "create",
            "DELETE": "delete",
        }
    },
    "ROBOT": {
        'KEY': 'robot',
        'VALUE': {
            "CREATE": "create",
            "UPDATE": "update",
            "DELETE": "delete",
            "LIST": "list",
            "READ": "read",
        }
    },
};

