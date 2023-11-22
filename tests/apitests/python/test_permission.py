import json
import random
import requests
import urllib3
import os
from urllib.parse import urlsplit


user_name = os.environ.get("USER_NAME")
password = os.environ.get("PASSWORD")
admin_user_name = os.environ.get("ADMIN_USER_NAME")
admin_password = os.environ.get("ADMIN_PASSWORD")
harbor_base_url = os.environ.get("HARBOR_BASE_URL")
resource = os.environ.get("RESOURCE")
ID_PLACEHOLDER = "(id)"
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

class Permission:


    def __init__(self, url, method, expect_status_code, payload=None, res_id_field=None, payload_id_field=None, id_from_header=False):
        self.url = url
        self.method = method
        self.expect_status_code = expect_status_code
        self.payload = payload
        self.res_id_field = res_id_field
        self.payload_id_field = payload_id_field if payload_id_field else res_id_field
        self.id_from_header = id_from_header


    def call(self):
        if ID_PLACEHOLDER in self.url:
            self.url = self.url.replace(ID_PLACEHOLDER, str(self.payload.get(self.payload_id_field)))
        response = requests.request(self.method, self.url, data=json.dumps(self.payload), verify=False, auth=(user_name, password), headers={"Content-Type": "application/json"})
        assert response.status_code == self.expect_status_code, "Failed to call the {} {}, expected status code is {}, but got {}, error msg is {}".format(self.method, self.url, self.expect_status_code, response.status_code, response.text)
        if self.res_id_field and self.payload_id_field and self.id_from_header == False:
            self.payload[self.payload_id_field] = int(json.loads(response.text)[self.res_id_field])
        elif self.res_id_field and self.payload_id_field and self.id_from_header == True:
            self.payload[self.payload_id_field] = int(response.headers["Location"].split("/")[-1])

resource_permissions = {}
# audit logs permissions start
list_audit_logs = Permission("{}/audit-logs".format(harbor_base_url), "GET", 200)
audit_log = [ list_audit_logs ]
resource_permissions["audit-log"] = audit_log
# audit logs permissions end

# preheat instance permissions start
preheat_instance_payload = {
    "name": "preheat_instance-{}".format(random.randint(1000, 9999)),
    "endpoint": "http://{}".format(random.randint(1000, 9999)),
    "enabled": False,
    "vendor": "dragonfly",
    "auth_mode": "NONE",
    "insecure": True
}
create_preheat_instance = Permission("{}/p2p/preheat/instances".format(harbor_base_url), "POST", 201, preheat_instance_payload)
list_preheat_instance = Permission("{}/p2p/preheat/instances".format(harbor_base_url), "GET", 200, preheat_instance_payload)
read_preheat_instance = Permission("{}/p2p/preheat/instances/{}".format(harbor_base_url, preheat_instance_payload["name"]), "GET", 200, preheat_instance_payload, "id")
update_preheat_instance = Permission("{}/p2p/preheat/instances/{}".format(harbor_base_url, preheat_instance_payload["name"]), "PUT", 200, preheat_instance_payload)
delete_preheat_instance = Permission("{}/p2p/preheat/instances/{}".format(harbor_base_url, preheat_instance_payload["name"]), "DELETE", 200, preheat_instance_payload)
ping_preheat_instance = Permission("{}/p2p/preheat/instances/ping".format(harbor_base_url), "POST", 500, preheat_instance_payload)
preheat_instances = [ create_preheat_instance, list_preheat_instance, read_preheat_instance, update_preheat_instance, delete_preheat_instance ]
resource_permissions["preheat-instance"] = preheat_instances
# preheat instance permissions end

# project permissions start
project_payload = {
	"metadata": {
		"public": "false"
	},
	"project_name": "project-{}".format(random.randint(1000, 9999)),
	"storage_limit": -1
}
create_project = Permission("{}/projects".format(harbor_base_url), "POST", 201, project_payload)
list_project = Permission("{}/projects".format(harbor_base_url), "GET", 200, project_payload)
project = [ create_project, list_project ]
resource_permissions["project"] = project
# project permissions end

# registry permissions start
registry_payload = {
    "insecure": False,
    "name": "registry-{}".format(random.randint(1000, 9999)),
    "type": "docker-hub",
    "url": "https://hub.docker.com"
}
create_registry = Permission("{}/registries".format(harbor_base_url), "POST", 201, registry_payload, "id", id_from_header=True)
list_registry = Permission("{}/registries".format(harbor_base_url), "GET", 200, registry_payload)
read_registry = Permission("{}/registries/{}".format(harbor_base_url, ID_PLACEHOLDER), "GET", 200, registry_payload, payload_id_field="id")
info_registry = Permission("{}/registries/{}/info".format(harbor_base_url, ID_PLACEHOLDER), "GET", 200, registry_payload, payload_id_field="id")
update_registry = Permission("{}/registries/{}".format(harbor_base_url, ID_PLACEHOLDER), "PUT", 200, registry_payload, payload_id_field="id")
delete_registry = Permission("{}/registries/{}".format(harbor_base_url, ID_PLACEHOLDER), "DELETE", 200, registry_payload, payload_id_field="id")
registry_ping_payload = {
    "insecure": False,
    "name": "registry-{}".format(random.randint(1000, 9999)),
    "type": "docker-hub",
    "url": "https://hub.docker.com"
}
ping_registry = Permission("{}/registries/ping".format(harbor_base_url), "POST", 200, registry_ping_payload)
registry = [ create_registry, list_registry, read_registry, info_registry, update_registry, delete_registry, ping_registry ]
resource_permissions["registry"] = registry
# registry permissions end

# replication-adapter permissions start
list_replication_adapters = Permission("{}/replication/adapters".format(harbor_base_url), "GET", 200)
list_replication_adapterinfos = Permission("{}/replication/adapterinfos".format(harbor_base_url), "GET", 200)
replication_adapter = [ list_replication_adapters, list_replication_adapterinfos ]
resource_permissions["replication-adapter"] = replication_adapter
# replication-adapter permissions end

# replication policy  permissions start
replication_registry_id = None
replication_registry_name = "replication-registry-{}".format(random.randint(1000, 9999))
if resource == "replication-policy":
    result = urlsplit(harbor_base_url)
    endpoint_URL = "{}://{}".format(result.scheme, result.netloc)
    replication_registry_payload = {
        "credential": {
            "access_key": admin_user_name,
            "access_secret": admin_password,
            "type": "basic"
        },
        "description": "",
        "insecure": True,
        "name": replication_registry_name,
        "type": "harbor",
        "url": endpoint_URL
    }
    response = requests.post("{}/registries".format(harbor_base_url), data=json.dumps(replication_registry_payload), verify=False, auth=(admin_user_name, admin_password), headers={"Content-Type": "application/json"})
    replication_registry_id = int(response.headers["Location"].split("/")[-1])
replication_policy_payload = {
    "name": "replication_policy_{}".format(random.randint(1000, 9999)),
    "src_registry": None,
    "dest_registry": {
        "id": replication_registry_id
    },
    "dest_namespace": "library",
    "dest_namespace_replace_count": 1,
    "trigger": {
        "type": "manual",
        "trigger_settings": {
            "cron": ""
        }
    },
    "filters": [
        {
            "type": "name",
            "value": "library/**"
        }
    ],
    "enabled": True,
    "deletion": False,
    "override": True,
    "speed": -1,
    "copy_by_chunk": False
}
create_replication_policy = Permission("{}/replication/policies".format(harbor_base_url), "POST", 201, replication_policy_payload, "id", id_from_header=True)
list_replication_policy = Permission("{}/replication/policies".format(harbor_base_url), "GET", 200, replication_policy_payload)
read_replication_policy = Permission("{}/replication/policies/{}".format(harbor_base_url, ID_PLACEHOLDER), "GET", 200, replication_policy_payload, payload_id_field="id")
update_replication_policy = Permission("{}/replication/policies/{}".format(harbor_base_url, ID_PLACEHOLDER), "PUT", 200, replication_policy_payload, payload_id_field="id")
delete_replication_policy = Permission("{}/replication/policies/{}".format(harbor_base_url, ID_PLACEHOLDER), "DELETE", 200, replication_policy_payload, payload_id_field="id")
replication_and_policy = [ create_replication_policy, list_replication_policy, read_replication_policy, update_replication_policy, delete_replication_policy ]
resource_permissions["replication-policy"] = replication_and_policy
# replication policy  permissions end

# replication permissions start
replication_policy_id = None
replication_policy_name = "replication-policy-{}".format(random.randint(1000, 9999))
if resource == "replication":
    result = urlsplit(harbor_base_url)
    endpoint_URL = "{}://{}".format(result.scheme, result.netloc)
    replication_registry_payload = {
        "credential": {
            "access_key": admin_user_name,
            "access_secret": admin_password,
            "type": "basic"
        },
        "description": "",
        "insecure": True,
        "name": "replication-registry-{}".format(random.randint(1000, 9999)),
        "type": "harbor",
        "url": endpoint_URL
    }
    response = requests.post("{}/registries".format(harbor_base_url), data=json.dumps(replication_registry_payload), verify=False, auth=(admin_user_name, admin_password), headers={"Content-Type": "application/json"})
    replication_registry_id = int(response.headers["Location"].split("/")[-1])
    replication_policy_payload = {
        "name": replication_policy_name,
        "src_registry": None,
        "dest_registry": {
            "id": replication_registry_id
        },
        "dest_namespace": "library",
        "dest_namespace_replace_count": 1,
        "trigger": {
            "type": "manual",
            "trigger_settings": {
                "cron": ""
            }
        },
        "filters": [
            {
                "type": "name",
                "value": "library/**"
            }
        ],
        "enabled": True,
        "deletion": False,
        "override": True,
        "speed": -1,
        "copy_by_chunk": False
    }
    response = requests.post("{}/replication/policies".format(harbor_base_url), data=json.dumps(replication_policy_payload), verify=False, auth=(admin_user_name, admin_password), headers={"Content-Type": "application/json"})
    replication_policy_id = int(response.headers["Location"].split("/")[-1])
replication_execution_payload = {
    "policy_id": replication_policy_id
}
create_replication_execution = Permission("{}/replication/executions".format(harbor_base_url), "POST", 201, replication_execution_payload, "id", id_from_header=True)
list_replication_execution = Permission("{}/replication/executions".format(harbor_base_url), "GET", 200, replication_execution_payload)
read_replication_execution = Permission("{}/replication/executions/{}".format(harbor_base_url, ID_PLACEHOLDER), "GET", 200, replication_execution_payload, payload_id_field="id")
stop_replication_execution = Permission("{}/replication/executions/{}".format(harbor_base_url, ID_PLACEHOLDER), "PUT", 200, replication_execution_payload, payload_id_field="id")
list_replication_execution_tasks = Permission("{}/replication/executions/{}/tasks".format(harbor_base_url, ID_PLACEHOLDER), "GET", 200, replication_execution_payload, payload_id_field="id")
read_replication_execution_task = Permission("{}/replication/executions/{}/tasks/{}".format(harbor_base_url, ID_PLACEHOLDER, 1), "GET", 404, replication_execution_payload, payload_id_field="id")
replication = [ create_replication_execution, list_replication_execution, read_replication_execution, stop_replication_execution, list_replication_execution_tasks, read_replication_execution_task ]
resource_permissions["replication"] = replication
# replication permissions end



def main():
    for permission in resource_permissions[resource]:
        print("=================================================")
        print("call: {} {}".format(permission.method, permission.url))
        print("payload: {}".format(json.dumps(permission.payload)))
        print("=================================================\n")
        permission.call()


if __name__ == "__main__":
    main()
