import os
import json
import argparse
import requests

from requests.packages.urllib3.exceptions import InsecureRequestWarning
requests.packages.urllib3.disable_warnings(InsecureRequestWarning)

parser = argparse.ArgumentParser(description='The script to generate data for harbor v1.4.0')
parser.add_argument('--endpoint', '-e', dest='endpoint', required=True, help='The endpoint to harbor')
parser.add_argument('--version', '-v', dest='version', required=False, help='The version to harbor')
args = parser.parse_args()

url = "https://"+args.endpoint+"/api/"
endpoint_url = "https://"+args.endpoint
print url

class HarborAPI:
    def create_project(self, project_name):
        body=dict(body={"project_name": ""+project_name+"", "metadata": {"public": "true"}})
        request(url+"projects", 'post', **body)

    def create_user(self, username):
        payload = {"username":username, "email":username+"@vmware.com", "password":"Harbor12345", "realname":username, "comment":"string"}
        body=dict(body=payload)
        request(url+"users", 'post', **body)

    def set_user_admin(self, user):
        r = request(url+"users?username="+user+"", 'get')
        userid = str(r.json()[0]['user_id'])
        if args.version == "1.6":
            body=dict(body={"sysadmin_flag": True})
        else:
            body=dict(body={"sysadmin_flag": 1})
        request(url+"users/"+userid+"/sysadmin", 'put', **body)

    def add_member(self, project, user, role):
        r = request(url+"projects?name="+project+"", 'get')
        projectid = str(r.json()[0]['project_id'])
        if args.version == "1.6":
            payload = {"member_user":{ "username": ""+user+""},"role_id": role}
        else:
            payload = {"roles": [role], "username":""+user+""}

        body=dict(body=payload)
        request(url+"projects/"+projectid+"/members", 'post', **body)

    def add_endpoint(self, endpointurl, endpointname, username, password, insecure):
        payload = {
            "credential":{
                "access_key":""+username+"",
                "access_secret":""+password+"",
                "type":"basic"
            },
            "insecure":insecure,
            "name":""+endpointname+"",
            "type":"harbor",
            "url":""+endpoint_url+""
        }
        body=dict(body=payload)
        print  body
        request(url+"/registries", 'post', **body)

    def add_replication_rule(self, project, target, trigger, rulename):
        r = request(url+"registries?name="+target+"", 'get')
        targetid = r.json()[0]['id']
        payload = {"name": ""+rulename+"", "deletion": False, "enabled": True, "description": "string", "dest_registry": {"id": targetid},"trigger": {"type": "manual"}}
        body=dict(body=payload)
        request(url+"replication/policies", 'post', **body)

    def update_project_setting(self, project, public, contenttrust, preventrunning, preventseverity, scanonpush):
        r = request(url+"projects?name="+project+"", 'get')
        projectid = str(r.json()[0]['project_id'])
        payload = {
            "project_name": ""+project+"",
            "metadata": {
                "public": public,
                "enable_content_trust": contenttrust,
                "prevent_vulnerable_images_from_running": preventrunning,
                "prevent_vulnerable_images_from_running_severity": preventseverity,
                "automatically_scan_images_on_push": scanonpush
            }
        }
        body=dict(body=payload)
        request(url+"projects/"+projectid+"", 'put', **body)

    def update_systemsetting(self, emailfrom, emailhost, emailport, emailuser, creation, selfreg, token):
        payload = {
            "auth_mode": "db_auth",
            "email_from": emailfrom,
            "email_host": emailhost,
            "email_port": emailport,
            "email_identity": "string",
            "email_username": emailuser,
            "email_ssl": True,
            "email_insecure": True,
            "project_creation_restriction": creation,
            "read_only": False,
            "self_registration": selfreg,
            "token_expiration": token,
            "scan_all_policy": {
                "type": "none",
                "parameter": {
                    "daily_time": 0
                }
            }
        }
        body=dict(body=payload)
        request(url+"configurations", 'put', **body)

    def update_repoinfo(self, reponame):
        payload = {"description": "testdescription"}
        body=dict(body=payload)
        request(url+"repositories/"+reponame+"", 'put', **body)

    def get_ca(self, target='/harbor/ca/ca.crt'):
        url = "https://" + args.endpoint + "/api/systeminfo/getcert"
        resp = request(url, 'get')
        try:
            ca_content = json.loads(resp.text)
        except ValueError:
            ca_content = resp.text
        ca_path = '/harbor/ca'
        if not os.path.exists(ca_path):
            try:
                os.makedirs(ca_path)
            except Exception, e:
                pass
        open(target, 'wb').write(ca_content)


def request(url, method, user = None, userp = None, **kwargs):
    if user is None:
        user = "admin"
    if userp is None:
        userp = "Harbor12345"
    kwargs.setdefault('headers', kwargs.get('headers', {}))
    kwargs['headers']['Accept'] = 'application/json'
    if 'body' in kwargs:
        kwargs['headers']['Content-Type'] = 'application/json'
        kwargs['data'] = json.dumps(kwargs['body'])
        del kwargs['body']

    resp = requests.request(method, url, verify=False, auth=(user, userp), **kwargs)
    if resp.status_code >= 400:
        raise Exception("[Exception Message] - {}".format(resp.text))
    return resp

with open("data.json") as f:
    data = json.load(f)

def pull_image(*image):
    for i in image:
        os.system("docker pull "+i)

def push_image(image, project):
    os.system("docker tag "+image+" "+args.endpoint+"/"+project+"/"+image)
    os.system("docker login "+args.endpoint+" -u Admin"+" -p Harbor12345")
    os.system("docker push "+args.endpoint+"/"+project+"/"+image)

def push_signed_image(image, project, tag):
    os.system("./sign_image.sh" + " " + args.endpoint + " " + project + " " + image + " " + tag)

def do_data_creation():
    harborAPI = HarborAPI()
    harborAPI.get_ca()

    for user in data["users"]:
        harborAPI.create_user(user["name"])

    for user in data["admin"]:
        harborAPI.set_user_admin(user["name"])

    for project in data["projects"]:
        harborAPI.create_project(project["name"])
        for member in project["member"]:
            harborAPI.add_member(project["name"], member["name"], member["role"])

    pull_image("busybox", "redis", "haproxy", "alpine", "httpd:2")
    push_image("busybox", data["projects"][0]["name"])
    push_signed_image("alpine", data["projects"][0]["name"], "latest")

    for endpoint in data["endpoint"]:
        harborAPI.add_endpoint(endpoint["url"], endpoint["name"], endpoint["user"], endpoint["pass"], True)
    for replicationrule in data["replicationrule"]:
        harborAPI.add_replication_rule(replicationrule["project"],
                                       replicationrule["endpoint"], replicationrule["trigger"],
                                       replicationrule["rulename"])
    for project in data["projects"]:
        harborAPI.update_project_setting(project["name"],
                                        project["configuration"]["public"],
                                        project["configuration"]["enable_content_trust"],
                                        project["configuration"]["prevent_vulnerable_images_from_running"],
                                        project["configuration"]["prevent_vlunerable_images_from_running_severity"],
                                        project["configuration"]["automatically_scan_images_on_push"])
    harborAPI.update_systemsetting(data["configuration"]["emailsetting"]["emailfrom"],
                                   data["configuration"]["emailsetting"]["emailserver"],
                                   float(data["configuration"]["emailsetting"]["emailport"]),
                                   data["configuration"]["emailsetting"]["emailuser"],
                                   data["configuration"]["projectcreation"],
                                   data["configuration"]["selfreg"],
                                   float(data["configuration"]["token"]))
do_data_creation()