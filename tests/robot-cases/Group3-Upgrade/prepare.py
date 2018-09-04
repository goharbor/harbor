import os
import json
import argparse
import requests

from requests.packages.urllib3.exceptions import InsecureRequestWarning
requests.packages.urllib3.disable_warnings(InsecureRequestWarning)

parser = argparse.ArgumentParser(description='The script to generate data for harbor v1.4.0')
parser.add_argument('--endpoint', '-e', dest='endpoint', required=True, help='The endpoint to harbor')
args = parser.parse_args() 

url = "https://"+args.endpoint+"/api/"
print url

class HarborAPI:

    def create_project(self, project_name):
        r = requests.post(url+"projects", auth=("admin", "Harbor12345"), json={"project_name": ""+project_name+"", "metadata": {"public": "true"}}, verify=False)
        print(r.status_code)
    
    def create_user(self, username):
        payload = {"username":username, "email":username+"@vmware.com", "password":"Harbor12345", "realname":username, "commment":"string"}
        r = requests.post(url+"users", auth=("admin", "Harbor12345"), json=payload, verify=False)
        print(r.status_code)
    
    def set_user_admin(self, user):
        r = requests.get(url+"users?username="+user+"", auth=("admin", "Harbor12345"), verify=False)
        userid = str(r.json()[0]['user_id'])
        r = requests.put(url+"users/"+userid+"/sysadmin", auth=("admin", "Harbor12345"), json={"has_admin_role": 1}, verify=False)
        print(r.status_code)
    
    def add_member(self, project, user, role):
        r = requests.get(url+"projects?name="+project+"", auth=("admin", "Harbor12345"), verify=False)
        projectid = str(r.json()[0]['project_id'])
        payload = {"roles": [role], "username":""+user+""}
        r = requests.post(url+"projects/"+projectid+"/members", auth=("admin", "Harbor12345"), json=payload, verify=False)
        print(r.status_code)
   
    def add_endpoint(self, endpointurl, endpointname, username, password, insecure):
        payload = {"endpoint": ""+endpointurl+"", "name": ""+endpointname+"", "username": ""+username+"", "password": ""+password+"", "insecure": insecure}
        r = requests.post(url+"targets", auth=("admin", "Harbor12345"), json=payload, verify=False)
        print(r.status_code)
    
    def add_replication_rule(self, project, target, trigger, rulename):
        r = requests.get(url+"projects?name="+project+"", auth=("admin", "Harbor12345"), verify=False)
        projectid = r.json()[0]['project_id']
        r = requests.get(url+"targets?name="+target+"", auth=("admin", "Harbor12345"), verify=False)
        targetid = r.json()[0]['id']
        payload = {"name": ""+rulename+"", "description": "string", "projects": [{"project_id": projectid,}], "targets": [{"id": targetid,}], "trigger": {"kind": ""+trigger+"", "schedule_param": {"type": "weekly", "weekday": 1, "offtime": 0}}}
        r = requests.post(url+"policies/replication", auth=("admin", "Harbor12345"), json=payload, verify=False)
        print(r.status_code)
    
    def update_project_setting(self, project, contenttrust, preventrunning, preventseverity, scanonpush):
        r = requests.get(url+"projects?name="+project+"", auth=("admin", "Harbor12345"), verify=False)
        projectid = str(r.json()[0]['project_id'])
        payload = {
            "project_name": ""+project+"",
            "metadata": {
                "public": "True",
                "enable_content_trust": contenttrust,
                "prevent_vulnerable_images_from_running": preventrunning,
                "prevent_vulnerable_images_from_running_severity": preventseverity,
                "automatically_scan_images_on_push": scanonpush
            }
        }
        r = requests.put(url+"projects/"+projectid+"", auth=("admin", "Harbor12345"), json=payload, verify=False)
        print(r.status_code)
    
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
            "verify_remote_cert": True,
            "scan_all_policy": {
                "type": "none",
                "parameter": {
                    "daily_time": 0
                }
            }
        }
        r = requests.put(url+"configurations", auth=("admin", "Harbor12345"), json=payload, verify=False)
        print(r.status_code)
    
    def update_repoinfo(self, reponame):
        r = requests.put(url+"repositories/"+reponame+"", auth=("admin", "Harbor12345"), json={"description": "testdescription"}, verify=False)
        print(r.status_code)

    def get_ca(self, target='/harbor/ca/ca.crt'):
        ca_content = request(args.endpoint, '/systeminfo/getcert', 'get', "admin", "Harbor12345")
        ca_path = '/harbor/ca'
        if not os.path.exists(ca_path):
            try:
                os.makedirs(ca_path)
            except Exception, e:
                pass
        open(target, 'wb').write(ca_content)

def request(harbor_endpoint, url, method, user, pwd, **kwargs):
    url = "https://" + harbor_endpoint + "/api" + url
    kwargs.setdefault('headers', kwargs.get('headers', {}))
    kwargs['headers']['Accept'] = 'application/json'
    if 'body' in kwargs:
        kwargs['headers']['Content-Type'] = 'application/json'
        kwargs['data'] = json.dumps(kwargs['body'])
        del kwargs['body']

    resp = requests.request(method, url, verify=False, auth=(user, pwd), **kwargs)
    if resp.status_code >= 400:
        raise Exception("Error: %s" % resp.text)
    try:
        body = json.loads(resp.text)
    except ValueError:
        body = resp.text
    return body

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
        harborAPI.add_endpoint(endpoint["url"], endpoint["name"], endpoint["user"], endpoint["pass"], False)
    
    for replicationrule in data["replicationrule"]:
        harborAPI.add_replication_rule(replicationrule["project"], 
                                       replicationrule["endpoint"], replicationrule["trigger"], 
                                       replicationrule["rulename"])
    
    for project in data["projects"]:
        harborAPI.update_project_setting(project["name"], 
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