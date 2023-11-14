import json
import random
import requests
import urllib3
import os


user_name = os.environ.get("USER_NAME")
password = os.environ.get("PASSWORD")
harbor_base_url = os.environ.get("HARBOR_BASE_URL")
resource = os.environ.get("RESOURCE")


class Permission:


    def __init__(self, url, method, expect_status_code, payload=None, res_id_field=None, payload_id_field=None):
        self.url = url
        self.method = method
        self.expect_status_code = expect_status_code
        self.payload = payload
        self.res_id_field = res_id_field
        self.payload_id_field = payload_id_field if payload_id_field else res_id_field


    def call(self):
        urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)
        response = None
        requests.get
        response = requests.request(self.method, self.url, data=json.dumps(self.payload), verify=False, auth=(user_name, password), headers={"Content-Type": "application/json"})
        print(response.text)
        assert response.status_code == self.expect_status_code, "Failed to call the {} {}, expected status code is {}, but got {}, error msg is {}".format(self.method, self.url, self.expect_status_code, response.status_code, response.text)
        if self.res_id_field and self.payload_id_field:
            self.payload[self.payload_id_field] = json.loads(response.text)[self.res_id_field]


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
resource_permissions = {
    "preheat-instance": preheat_instances
}
# preheat instance permissions end


def main():
    for permission in resource_permissions[resource]:
        permission.call()


if __name__ == "__main__":
    main()
