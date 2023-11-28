import copy
import json
import time
import requests
import urllib3
import os

admin_name = os.environ.get("ADMIN_NAME")
admin_password = os.environ.get("ADMIN_PASSWORD")
user_name = os.environ.get("USER_NAME")
password = os.environ.get("PASSWORD")
harbor_base_url = os.environ.get("HARBOR_BASE_URL")
resource = os.environ.get("RESOURCE")
project_id = os.environ.get("PROJECT_ID")
project_name = os.environ.get("PROJECT_NAME")
# the source artifact should belong to the provided project, e.g. "nginx"
source_artifact_name = os.environ.get("SOURCE_ARTIFACT_NAME")
# the source artifact tag should belong to the provided project, e.g. "latest"
source_artifact_tag = os.environ.get("SOURCE_ARTIFACT_TAG")
id_or_name = None

urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)


class Permission:

    def __init__(self, url, method, expect_status_code, payload=None, need_id_or_name=False, res_id_field=None,
                 payload_id_field=None):
        self.url = url
        self.method = method
        self.expect_status_code = expect_status_code
        self.payload = payload
        self.res_id_field = res_id_field
        self.need_id_or_name = need_id_or_name
        self.payload_id_field = payload_id_field if payload_id_field else res_id_field

    def call(self):
        global id_or_name
        url = self.url
        if self.need_id_or_name:
            url = self.url.format(id_or_name)
        response = requests.request(self.method, url, data=json.dumps(self.payload), verify=False,
                                    auth=(user_name, password), headers={
                "Content-Type": "application/json"
            })
        print("response: {}".format(response.text))
        assert response.status_code == self.expect_status_code, ("Failed to call the {} {}, expected status code is {"
                                                                 "}, but got {}, error msg is {}").format(
            self.method, self.url, self.expect_status_code, response.status_code, response.text)
        if self.res_id_field and self.payload_id_field and len(json.loads(response.text)) > 0:
            id_or_name = json.loads(response.text)[0][self.res_id_field]


# Project permissions:
# 1. Resource: label, actions: ['read', 'create', 'update', 'delete', 'list']
label_payload = {
    "color": "#FFFFFF",
    "description": "Just for testing",
    "name": "label-name-{}".format(int(round(time.time() * 1000))),
    "project_id": int(project_id),
    "scope": "p",
    "id": None
}
create_label = Permission("{}/labels".format(harbor_base_url), "POST", 201, label_payload)
list_label = Permission("{}/labels?scope=p&project_id={}".format(harbor_base_url, project_id), "GET", 200,
                        label_payload, False, "id", "id")
read_label = Permission("{}/labels/{}".format(harbor_base_url, "{}"), "GET", 200, label_payload, True)
label_payload_for_update = copy.deepcopy(label_payload)
label_payload_for_update["description"] = "For update"
update_label = Permission("{}/labels/{}".format(harbor_base_url, "{}"), "PUT", 200, label_payload_for_update, True)
delete_label = Permission("{}/labels/{}".format(harbor_base_url, "{}"), "DELETE", 200, label_payload, True)

# 2. Resource: project, actions: ['read', 'update', 'delete']
project_payload_for_update = {"project_name": "test", "metadata": {"public": "false"}, "storage_limit": -1}
read_project = Permission("{}/projects/{}".format(harbor_base_url, project_id), "GET", 200, project_payload_for_update,
                          False)
update_project = Permission("{}/projects/{}".format(harbor_base_url, project_id), "PUT", 200,
                            project_payload_for_update, False)
delete_project = Permission("{}/projects/{}".format(harbor_base_url, project_id), "DELETE", 412,
                            project_payload_for_update, False)
deletable_project = Permission("{}/projects/{}/_deletable".format(harbor_base_url, project_id), "GET", 200,
                               project_payload_for_update, False)

# 3. Resource: metadata   actions: ['read', 'list', 'create', 'update', 'delete'],
metadata_payload = {
    "auto_scan": "true"
}
create_metadata = Permission("{}/projects/{}/metadatas".format(harbor_base_url, project_id), "POST", 200,
                             metadata_payload, False)
list_metadata = Permission("{}/projects/{}/metadatas".format(harbor_base_url, project_id), "GET", 200, metadata_payload,
                           False, )
read_metadata = Permission("{}/projects/{}/metadatas/auto_scan".format(harbor_base_url, project_id), "GET", 200,
                           metadata_payload, False)
metadata_payload_for_update = {
    "auto_scan": "false"
}
update_metadata = Permission("{}/projects/{}/metadatas/auto_scan".format(harbor_base_url, project_id), "PUT", 200,
                             metadata_payload_for_update, False)
delete_metadata = Permission("{}/projects/{}/metadatas/auto_scan".format(harbor_base_url, project_id), "DELETE", 200,
                             metadata_payload, False)

# 4. Resource: repository  actions: ['read', 'list', 'update', 'delete', 'pull', 'push']
# note: pull and push are for docker cli,  no API needs them
list_repo = Permission("{}/projects/{}/repositories".format(harbor_base_url, project_name), "GET", 200)
read_repo = Permission("{}/projects/{}/repositories/does_not_exist".format(harbor_base_url, project_name), "GET", 404)
repo_payload_for_update = {
}
update_repo = Permission("{}/projects/{}/repositories/does_not_exist".format(harbor_base_url, project_name), "PUT", 404,
                         repo_payload_for_update)
delete_repo = Permission("{}/projects/{}/repositories/does_not_exist".format(harbor_base_url, project_name), "DELETE",
                         404)

# 5. Resource artifact   actions: ['read', 'list', 'create', 'delete'],
list_artifact = Permission("{}/projects/{}/repositories/does_not_exist/artifacts".format(harbor_base_url, project_name),
                           "GET", 200)
read_artifact = Permission(
    "{}/projects/{}/repositories/does_not_exist/artifacts/reference_does_not_exist".format(harbor_base_url,
                                                                                           project_name), "GET", 404)
copy_artifact = Permission(
    "{}/projects/{}/repositories/target_repo/artifacts?from={}/{}:{}".format(harbor_base_url, project_name,
                                                                             project_name, source_artifact_name,
                                                                             source_artifact_tag), "POST", 201)
delete_artifact = Permission(
    "{}/projects/{}/repositories/target_repo/artifacts/{}".format(harbor_base_url, project_name, source_artifact_tag),
    "DELETE", 200)

# 6. Resource scan      actions: ['read', 'create', 'stop']
create_scan = Permission(
    "{}/projects/{}/repositories/{}/artifacts/{}/scan".format(harbor_base_url, project_name, source_artifact_name,
                                                              source_artifact_tag), "POST", 202)
stop_scan = Permission(
    "{}/projects/{}/repositories/{}/artifacts/{}/scan/stop".format(harbor_base_url, project_name, source_artifact_name,
                                                                   source_artifact_tag), "POST", 202)
read_scan = Permission(
    "{}/projects/{}/repositories/{}/artifacts/{}/scan/0/log".format(harbor_base_url, project_name, source_artifact_name,
                                                                    source_artifact_tag), "get", 404)

# 7. Resource tag      actions: ['list', 'create', 'delete']
tag_payload = {
    "name": "test-{}".format(int(round(time.time() * 1000)))
}
create_tag = Permission(
    "{}/projects/{}/repositories/{}/artifacts/{}/tags".format(harbor_base_url, project_name, source_artifact_name,
                                                              source_artifact_tag), "POST", 201, tag_payload)
list_tag = Permission(
    "{}/projects/{}/repositories/{}/artifacts/{}/tags".format(harbor_base_url, project_name, source_artifact_name,
                                                              source_artifact_tag), "GET", 200)
delete_tag = Permission(
    "{}/projects/{}/repositories/{}/artifacts/{}/tags/tag_name_does_not_exist".format(harbor_base_url, project_name,
                                                                                      source_artifact_name,
                                                                                      source_artifact_tag), "DELETE",
    404)

# 8. Resource accessory  actions: ['list']
list_accessory = Permission(
    "{}/projects/{}/repositories/{}/artifacts/{}/accessories".format(harbor_base_url, project_name,
                                                                     source_artifact_name, source_artifact_tag), "GET",
    200)

# 9. Resource artifact-addition    actions: ['read']
read_artifact_addition_vul = Permission(
    "{}/projects/{}/repositories/{}/artifacts/{}/additions/vulnerabilities".format(harbor_base_url, project_name,
                                                                                   source_artifact_name,
                                                                                   source_artifact_tag), "GET", 200)
read_artifact_addition_dependencies = Permission(
    "{}/projects/{}/repositories/{}/artifacts/{}/additions/dependencies".format(harbor_base_url, project_name,
                                                                                source_artifact_name,
                                                                                source_artifact_tag), "GET", 400)

# 10. Resource artifact-label     actions: ['create', 'delete'],
artifact_label_payload = copy.deepcopy(label_payload)
artifact_label_payload["description"] = "Add label to an artifact"
add_label_to_artifact = Permission(
    "{}/projects/{}/repositories/{}/artifacts/{}/labels".format(harbor_base_url, project_name, source_artifact_name,
                                                                source_artifact_tag), "POST", 404,
    artifact_label_payload)
delete_artifact_label = Permission(
    "{}/projects/{}/repositories/{}/artifacts/{}/labels/0".format(harbor_base_url, project_name, source_artifact_name,
                                                                  source_artifact_tag), "DELETE", 404,
    artifact_label_payload)

# 11. Resource scanner           actions: ['create', 'read']
update_project_scanner = Permission("{}/projects/{}/scanner".format(harbor_base_url, project_id), "PUT", 200,
                                    {"uuid": "faked_uuid"})
read_project_scanner = Permission("{}/projects/{}/scanner".format(harbor_base_url, project_id), "GET", 200)
read_project_scanner_candidates = Permission("{}/projects/{}/scanner/candidates".format(harbor_base_url, project_id),
                                             "GET", 200)

# 12. Resource preheat-policy   actions: ['read', 'list', 'create', 'update', 'delete']
create_preheat_policy = Permission("{}/projects/{}/preheat/policies".format(harbor_base_url, project_name), "POST", 500,
                                   {})
list_preheat_policy = Permission("{}/projects/{}/preheat/policies".format(harbor_base_url, project_name), "GET", 200)
read_preheat_policy = Permission(
    "{}/projects/{}/preheat/policies/policy_name_does_not_exist".format(harbor_base_url, project_name), "GET", 404)
update_preheat_policy = Permission(
    "{}/projects/{}/preheat/policies/policy_name_does_not_exist".format(harbor_base_url, project_name), "PUT", 500)
delete_preheat_policy = Permission(
    "{}/projects/{}/preheat/policies/policy_name_does_not_exist".format(harbor_base_url, project_name), "DELETE", 404)

# 13. Resource immutable-tag   actions: ['list', 'create', 'update', 'delete']
immutable_tag_rule_payload = {
    "disabled": False,
    "scope_selectors": {
        "repository": [{"kind": "doublestar", "decoration": "repoMatches",
                        "pattern": "{}".format(int(round(time.time() * 1000)))}]},
    "tag_selectors": [
        {"kind": "doublestar", "decoration": "matches", "pattern": "{}".format(int(round(time.time() * 1000)))}],
}
create_immutable_tag_rule = Permission("{}/projects/{}/immutabletagrules".format(harbor_base_url, project_id), "POST",
                                       201,
                                       immutable_tag_rule_payload)
list_immutable_tag_rule = Permission("{}/projects/{}/immutabletagrules".format(harbor_base_url, project_id), "GET", 200)
update_immutable_tag_rule = Permission("{}/projects/{}/immutabletagrules/0".format(harbor_base_url, project_id), "PUT",
                                       404)
delete_immutable_tag_rule = Permission("{}/projects/{}/immutabletagrules/0".format(harbor_base_url, project_id),
                                       "DELETE", 404)

# 14. Resource tag-retention   actions: ['read', 'list', 'create', 'update', 'delete']
tag_retention_rule_payload = {
    "algorithm": "or",
    "rules": [
        {
            "disabled": False,
            "action": "retain",
            "scope_selectors": {
                "repository": [
                    {
                        "kind": "doublestar",
                        "decoration": "repoMatches",
                        "pattern": "**"
                    }
                ]
            },
            "tag_selectors": [
                {
                    "kind": "doublestar",
                    "decoration": "matches",
                    "pattern": "**",
                    "extras": "{\"untagged\":true}"
                }
            ],
            "params": {},
            "template": "always"
        }
    ],
    "trigger": {
        "kind": "Schedule",
        "references": {},
        "settings": {
            "cron": ""
        }
    },
    "scope": {
        "level": "project",
        "ref": int(project_id)
    }
}


def get_retention_id() -> str:
    # create retention rule fist
    # this request can be failed(retention rule existed) or succeeded, but we can finally get the retention id
    requests.request("POST", "{}/retentions".format(harbor_base_url),
                                data=json.dumps(tag_retention_rule_payload), verify=False,
                                auth=(admin_name, admin_password), headers={"Content-Type": "application/json"})
    response1 = requests.request("GET", "{}/projects/{}/metadatas/retention_id".format(harbor_base_url, project_id),
                                 data=None, verify=False,
                                 auth=(admin_name, admin_password), headers={"Content-Type": "application/json"})
    retention_id = project_id
    if "retention_id" in json.loads(response1.text):
        retention_id = json.loads(response1.text)["retention_id"]
    return retention_id


# because get_retention_id() has been called,  so the expected status code is 400
create_tag_retention_rule = Permission("{}/retentions".format(harbor_base_url), "POST",
                                       400,
                                       tag_retention_rule_payload)

update_retention_payload = copy.deepcopy(tag_retention_rule_payload)
update_retention_payload["rules"][0]["disabled"] = True
read_tag_retention = Permission("{}/retentions/{}".format(harbor_base_url, get_retention_id()), "GET", 200)
update_tag_retention = Permission("{}/retentions/{}".format(harbor_base_url, get_retention_id()), "PUT", 200,
                                  update_retention_payload)
delete_tag_retention = Permission("{}/retentions/{}".format(harbor_base_url, get_retention_id()), "DELETE", 200)
execute_tag_retention = Permission("{}/retentions/{}/executions".format(harbor_base_url, get_retention_id()), "POST",
                                   201)
list_tag_retention_execution = Permission("{}/retentions/{}/executions".format(harbor_base_url, get_retention_id()),
                                          "GET",
                                          200)
stop_tag_retention = Permission("{}/retentions/{}/executions/0".format(harbor_base_url, get_retention_id()), "PATCH",
                                404,
                                {"action": "stop"})
list_tag_retention_tasks = Permission("{}/retentions/{}/executions/0/tasks".format(harbor_base_url, get_retention_id()),
                                      "GET", 404)
read_tag_retention_tasks = Permission(
    "{}/retentions/{}/executions/0/tasks/0".format(harbor_base_url, get_retention_id()),
    "GET", 404)

# 15. Resource log   actions: ['list']
list_log = Permission("{}/projects/{}/logs".format(harbor_base_url, project_name), "GET", 200)

# 16. Resource notification-policy    actions: ['read', 'list', 'create', 'update', 'delete']
webhook_payload = {
    "name": "webhook-{}".format(int(round(time.time() * 1000))),
    "description": "Just for test",
    "project_id": int(project_id),
    "targets": [
        {
            "type": "http",
            "address": "http://test.com",
            "skip_cert_verify": True,
            "payload_format": "CloudEvents"
        }
    ],
    "event_types": [
        "PUSH_ARTIFACT"
    ],
    "enabled": True
}

create_webhook = Permission("{}/projects/{}/webhook/policies".format(harbor_base_url, project_id), "POST",
                            201,
                            webhook_payload)
list_webhook = Permission("{}/projects/{}/webhook/policies".format(harbor_base_url, project_id), "GET",
                          200)
read_webhook = Permission("{}/projects/{}/webhook/policies/0".format(harbor_base_url, project_id), "GET",
                          404)
update_webhook = Permission("{}/projects/{}/webhook/policies/0".format(harbor_base_url, project_id), "PUT",
                            404, {})
delete_webhook = Permission("{}/projects/{}/webhook/policies/0".format(harbor_base_url, project_id), "DELETE",
                            404)

list_webhook_executions = Permission("{}/projects/{}/webhook/policies/0/executions".format(harbor_base_url, project_id),
                                     "GET", 404)
list_webhook_executions_tasks = Permission(
    "{}/projects/{}/webhook/policies/0/executions/0/tasks".format(harbor_base_url, project_id), "GET", 404)
read_webhook_executions_tasks = Permission(
    "{}/projects/{}/webhook/policies/0/executions/0/tasks/0/log".format(harbor_base_url, project_id), "GET", 404)
list_webhook_events = Permission("{}/projects/{}/webhook/events".format(harbor_base_url, project_id), "GET", 200)

resource_permissions = {
    "label": [create_label, list_label, read_label, update_label, delete_label],
    "project": [read_project, update_project, deletable_project, delete_project],
    "metadata": [create_metadata, list_metadata, read_metadata, update_metadata, delete_metadata],
    "repository": [list_repo, read_repo, update_repo, delete_repo],
    "artifact": [list_artifact, read_artifact, copy_artifact, delete_artifact],
    "scan": [create_scan, stop_scan, read_scan],
    "tag": [create_tag, list_tag, delete_tag],
    "accessory": [list_accessory],
    "artifact-addition": [read_artifact_addition_vul, read_artifact_addition_dependencies],
    "artifact-label": [add_label_to_artifact, delete_artifact_label],
    "scanner": [update_project_scanner, read_project_scanner, read_project_scanner_candidates],
    "preheat-policy": [create_preheat_policy, list_preheat_policy, read_preheat_policy, update_preheat_policy,
                       delete_preheat_policy],
    "immutable-tag": [create_immutable_tag_rule, list_immutable_tag_rule, update_immutable_tag_rule,
                      delete_immutable_tag_rule],
    "tag-retention": [create_tag_retention_rule, read_tag_retention, update_tag_retention, execute_tag_retention,
                      list_tag_retention_execution, stop_tag_retention, list_tag_retention_tasks,
                      read_tag_retention_tasks, delete_tag_retention],
    "log": [list_log],
    "notification-policy": [create_webhook, list_webhook, read_webhook, update_webhook, delete_webhook,
                            list_webhook_executions, list_webhook_executions_tasks, read_webhook_executions_tasks,
                            list_webhook_events]
}


def main():
    for permission in resource_permissions[resource]:
        print("=================================================")
        print("call: {} {}".format(permission.method, permission.url))
        print("payload: {}".format(json.dumps(permission.payload)))
        print("=================================================\n")
        permission.call()


if __name__ == "__main__":
    main()
