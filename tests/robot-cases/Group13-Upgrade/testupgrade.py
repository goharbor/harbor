import os
import json
import argparse
import requests

#usage: testupgrade.py host version
parser = argparse.ArgumentParser()
parser.add_argument('x')
#
parser.add_argument('y', type=float)

parser.add_argument('--https', action="store_true", default=False)

args = parser.parse_args()

if args.https:
    protocol = "https"
else:
    protocol = "http" 

version = args.y
host = args.x

url = ""+protocol+"://"+host+"/api/"

class Vonedotone:
    def createproject(project_name):
        r = requests.post(url+"projects", auth=("admin", "Harbor12345"), json={"project_name": ""+project_name+"", "metadata": {"public": "true"}}, verify=False)
        print(r.status_code)
    
    def createuser(username):
        payload = {"username":""+username+"", "email":""+username+"@vmware.com", "password":"Harbor12345", "realname":""+username+"", "commment":"string"}
        r = requests.post(url+"users", auth=("admin", "Harbor12345"), json=payload, verify=False)
        print(r.status_code)
    
    def setuseradmin(user):
        r = requests.get(url+"users?username="+user+"", auth=("admin", "Harbor12345"), verify=False)
        userid = str(r.json()[0]['user_id'])
        r = requests.put(url+"users/"+userid+"/sysadmin", auth=("admin", "Harbor12345"), json={"has_admin_role": 1}, verify=False)
        print(r.status_code)
    
    def addmember(project, user, role):
        r = requests.get(url+"projects?name="+project+"", auth=("admin", "Harbor12345"), verify=False)
        projectid = str(r.json()[0]['project_id'])
        payload = {"roles": [role], "username":""+user+""}
        r = requests.post(url+"projects/"+projectid+"/members", auth=("admin", "Harbor12345"), json=payload, verify=False)
        print(r.status_code)
    
    def addendpoint(endpointurl, endpointname, username, password, insecure):
        payload = {"endpoint": ""+endpointurl+"", "name": ""+endpointname+"", "username": ""+username+"", "password": ""+password+"", "insecure": insecure}
        r = requests.post(url+"targets", auth=("admin", "Harbor12345"), json=payload, verify=False)
        print(r.status_code)
    
    def addreplicationrule(project, target, rulename, enable):
        r = requests.get(url+"projects?name="+project+"", auth=("admin", "Harbor12345"), verify=False)
        projectid = r.json()[0]['project_id']
        r = requests.get(url+"targets?name="+target+"", auth=("admin", "Harbor12345"), verify=False)
        targetid = r.json()[0]['id']
        payload = {"project_id": projectid, "target_id": targetid, "name": ""+rulename+"", "enabled": enable}
        r = requests.post(url+"policies/replication", auth=("admin", "Harbor12345"), json=payload, verify=False)
        print(r.status_code)
    
    def updateprojectsetting(project, public):
        r = requests.get(url+"projects?name="+project+"", auth=("admin", "Harbor12345"), verify=False)
        projectid = str(r.json()[0]['project_id'])
        payload = {
            "public": public
            }
        r = requests.put(url+"projects/"+projectid+"/publicity", auth=("admin", "Harbor12345"), json=payload, verify=False)
        print(r.status_code)
    
    def updatesystemsetting(emailfrom, emailhost, emailport, emailuser, creation):
        payload = {
            "auth_mode": "db_auth",
            "email_from": emailfrom,
            "email_host": emailhost,
            "email_port": emailport,
            "email_identity": "string",
            "email_username": emailuser,
            "email_ssl": "0",
            "project_creation_restriction": creation,
            "self_registration": "0",
            "token_expiration": "10",
            "verify_remote_cert": "0"
        }
        r = requests.put(url+"configurations", auth=("admin", "Harbor12345"), json=payload)
        print(r.status_code)


class Vonedottwo:
    def createproject(project_name):
        r = requests.post(url+"projects", auth=("admin", "Harbor12345"), json={"project_name": ""+project_name+"", "metadata": {"public": "true"}}, verify=False)
        print(r.status_code)
    
    def createuser(username):
        payload = {"username":""+username+"", "email":""+username+"@vmware.com", "password":"Harbor12345", "realname":""+username+"", "commment":"string"}
        r = requests.post(url+"users", auth=("admin", "Harbor12345"), json=payload, verify=False)
        print(r.status_code)
    
    def setuseradmin(user):
        r = requests.get(url+"users?username="+user+"", auth=("admin", "Harbor12345"), verify=False)
        userid = str(r.json()[0]['user_id'])
        r = requests.put(url+"users/"+userid+"/sysadmin", auth=("admin", "Harbor12345"), json={"has_admin_role": 1}, verify=False)
        print(r.status_code)
    
    def addmember(project, user, role):
        r = requests.get(url+"projects?name="+project+"", auth=("admin", "Harbor12345"), verify=False)
        projectid = str(r.json()[0]['project_id'])
        payload = {"roles": [role], "username":""+user+""}
        r = requests.post(url+"projects/"+projectid+"/members", auth=("admin", "Harbor12345"), json=payload, verify=False)
        print(r.status_code)
    
    def addendpoint(endpointurl, endpointname, username, password, insecure):
        payload = {"endpoint": ""+endpointurl+"", "name": ""+endpointname+"", "username": ""+username+"", "password": ""+password+"", "insecure": insecure}
        r = requests.post(url+"targets", auth=("admin", "Harbor12345"), json=payload, verify=False)
        print(r.status_code)
    
    def addreplicationrule(project, target, rulename, enable):
        r = requests.get(url+"projects?name="+project+"", auth=("admin", "Harbor12345"), verify=False)
        projectid = r.json()[0]['project_id']
        r = requests.get(url+"targets?name="+target+"", auth=("admin", "Harbor12345"), verify=False)
        targetid = r.json()[0]['id']
        payload = {"project_id": projectid, "target_id": targetid, "name": ""+rulename+"", "enabled": enable}
        r = requests.post(url+"policies/replication", auth=("admin", "Harbor12345"), json=payload, verify=False)
        print(r.status_code)
    
    def updateprojectsetting(project, public):
        r = requests.get(url+"projects?name="+project+"", auth=("admin", "Harbor12345"), verify=False)
        projectid = str(r.json()[0]['project_id'])
        payload = {
            "public": public
            }
        r = requests.put(url+"projects/"+projectid+"/publicity", auth=("admin", "Harbor12345"), json=payload, verify=False)
        print(r.status_code)
    
    def updatesystemsetting(emailfrom, emailhost, emailport, emailuser, creation, selfreg, token):
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
        r = requests.put(url+"configurations", auth=("admin", "Harbor12345"), json=payload)
        print(r.status_code)

class Vonedotthree:
    def createproject(project_name):
        r = requests.post(url+"projects", auth=("admin", "Harbor12345"), json={"project_name": ""+project_name+"", "metadata": {"public": "true"}}, verify=False)
        print(r.status_code)
    
    def createuser(username):
        payload = {"username":""+username+"", "email":""+username+"@vmware.com", "password":"Harbor12345", "realname":""+username+"", "commment":"string"}
        r = requests.post(url+"users", auth=("admin", "Harbor12345"), json=payload, verify=False)
        print(r.status_code)
    
    def setuseradmin(user):
        r = requests.get(url+"users?username="+user+"", auth=("admin", "Harbor12345"), verify=False)
        userid = str(r.json()[0]['user_id'])
        r = requests.put(url+"users/"+userid+"/sysadmin", auth=("admin", "Harbor12345"), json={"has_admin_role": 1}, verify=False)
        print(r.status_code)
    
    def addmember(project, user, role):
        r = requests.get(url+"projects?name="+project+"", auth=("admin", "Harbor12345"), verify=False)
        projectid = str(r.json()[0]['project_id'])
        payload = {"roles": [role], "username":""+user+""}
        r = requests.post(url+"projects/"+projectid+"/members", auth=("admin", "Harbor12345"), json=payload, verify=False)
        print(r.status_code)
    
    def addendpoint(endpointurl, endpointname, username, password, insecure):
        payload = {"endpoint": ""+endpointurl+"", "name": ""+endpointname+"", "username": ""+username+"", "password": ""+password+"", "insecure": insecure}
        r = requests.post(url+"targets", auth=("admin", "Harbor12345"), json=payload, verify=False)
        print(r.status_code)
    
    def addreplicationrule(project, target, rulename, enable):
        r = requests.get(url+"projects?name="+project+"", auth=("admin", "Harbor12345"), verify=False)
        projectid = r.json()[0]['project_id']
        r = requests.get(url+"targets?name="+target+"", auth=("admin", "Harbor12345"), verify=False)
        targetid = r.json()[0]['id']
        payload = {"project_id": projectid, "target_id": targetid, "name": ""+rulename+"", "enabled": enable}
        r = requests.post(url+"policies/replication", auth=("admin", "Harbor12345"), json=payload, verify=False)
        print(r.status_code)
    
    def updateprojectsetting(project, contenttrust, preventrunning, preventseverity, scanonpush):
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
    
    def updatesystemsetting(emailfrom, emailhost, emailport, emailuser, creation, selfreg, token):
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
        r = requests.put(url+"configurations", auth=("admin", "Harbor12345"), json=payload)
        print(r.status_code)
    
    def updaterepoinfo(reponame):
        r = requests.put(url+"repositories/"+reponame+"", auth=("admin", "Harbor12345"), json={"description": "testdescription"}, verify=False)
        print(r.status_code)


class Vonedotfour:
    def createproject(project_name):
        r = requests.post(url+"projects", auth=("admin", "Harbor12345"), json={"project_name": ""+project_name+"", "metadata": {"public": "true"}}, verify=False)
        print(r.status_code)
    
    def createuser(username):
        payload = {"username":""+username+"", "email":""+username+"@vmware.com", "password":"Harbor12345", "realname":""+username+"", "commment":"string"}
        r = requests.post(url+"users", auth=("admin", "Harbor12345"), json=payload, verify=False)
        print(r.status_code)
    
    def setuseradmin(user):
        r = requests.get(url+"users?username="+user+"", auth=("admin", "Harbor12345"), verify=False)
        userid = str(r.json()[0]['user_id'])
        r = requests.put(url+"users/"+userid+"/sysadmin", auth=("admin", "Harbor12345"), json={"has_admin_role": 1}, verify=False)
        print(r.status_code)
    
    def addmember(project, user, role):
        r = requests.get(url+"projects?name="+project+"", auth=("admin", "Harbor12345"), verify=False)
        projectid = str(r.json()[0]['project_id'])
        payload = {"roles": [role], "username":""+user+""}
        r = requests.post(url+"projects/"+projectid+"/members", auth=("admin", "Harbor12345"), json=payload, verify=False)
        print(r.status_code)
   
    def addendpoint(endpointurl, endpointname, username, password, insecure):
        payload = {"endpoint": ""+endpointurl+"", "name": ""+endpointname+"", "username": ""+username+"", "password": ""+password+"", "insecure": insecure}
        r = requests.post(url+"targets", auth=("admin", "Harbor12345"), json=payload, verify=False)
        print(r.status_code)
    
    def addreplicationrule(project, target, trigger, rulename):
        r = requests.get(url+"projects?name="+project+"", auth=("admin", "Harbor12345"), verify=False)
        projectid = r.json()[0]['project_id']
        r = requests.get(url+"targets?name="+target+"", auth=("admin", "Harbor12345"), verify=False)
        targetid = r.json()[0]['id']
        payload = {"name": ""+rulename+"", "description": "string", "projects": [{"project_id": projectid,}], "targets": [{"id": targetid,}], "trigger": {"kind": ""+trigger+"", "schedule_param": {"type": "weekly", "weekday": 1, "offtime": 0}}}
        r = requests.post(url+"policies/replication", auth=("admin", "Harbor12345"), json=payload, verify=False)
        print(r.status_code)
    
    def updateprojectsetting(project, contenttrust, preventrunning, preventseverity, scanonpush):
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
    
    def updatesystemsetting(emailfrom, emailhost, emailport, emailuser, creation, selfreg, token):
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
        r = requests.put(url+"configurations", auth=("admin", "Harbor12345"), json=payload)
        print(r.status_code)
    
    def updaterepoinfo(reponame):
        r = requests.put(url+"repositories/"+reponame+"", auth=("admin", "Harbor12345"), json={"description": "testdescription"}, verify=False)
        print(r.status_code)


class Vonedotfive:
    def createproject(project_name):
        r = requests.post(url+"projects", auth=("admin", "Harbor12345"), json={"project_name": ""+project_name+"", "metadata": {"public": "true"}}, verify=False)
        print(r.status_code)
    
    def createuser(username):
        payload = {"username":""+username+"", "email":""+username+"@vmware.com", "password":"Harbor12345", "realname":""+username+"", "commment":"string"}
        r = requests.post(url+"users", auth=("admin", "Harbor12345"), json=payload, verify=False)
        print(r.status_code)
    
    def setuseradmin(user):
        r = requests.get(url+"users?username="+user+"", auth=("admin", "Harbor12345"), verify=False)
        userid = str(r.json()[0]['user_id'])
        r = requests.put(url+"users/"+userid+"/sysadmin", auth=("admin", "Harbor12345"), json={"has_admin_role": 1}, verify=False)
        print(r.status_code)
    
    def addmember(project, user, role):
        r = requests.get(url+"projects?name="+project+"", auth=("admin", "Harbor12345"), verify=False)
        projectid = str(r.json()[0]['project_id'])
        payload = {"role_id":role, "member_user":{"username":""+user+""}}
        r = requests.post(url+"projects/"+projectid+"/members", auth=("admin", "Harbor12345"), json=payload, verify=False)
        print(r.status_code)
    
    #def addlabeltotag(project, tag, label):
    #    r = requests.put()
    
    def addsyslabel(labelname):
        payload = {"name": ""+labelname+"", "description":"string", "color":"string", "scope":"g"}
        r = requests.post(url+"labels", auth=("admin", "Harbor12345"), json=payload, verify=False)
        print(r.status_code)
    
    def addprojectlabel(project, label):
        r = requests.get(url+"projects?name="+project+"", auth=("admin", "Harbor12345"), verify=False)
        projectid = r.json()[0]['project_id']
        payload = {"name":""+label+"", "description": "string", "color": "string", "scope": "p", "project_id": projectid}
        r = requests.post(url+"labels", auth=("admin", "Harbor12345"), json=payload, verify=False)
        print(r.status_code)
    
    def addendpoint(endpointurl, endpointname, username, password, insecure):
        payload = {"endpoint": ""+endpointurl+"", "name": ""+endpointname+"", "username": ""+username+"", "password": ""+password+"", "insecure": insecure}
        r = requests.post(url+"targets", auth=("admin", "Harbor12345"), json=payload, verify=False)
        print(r.status_code)
    
    def addreplicationrule(project, target, trigger, rulename):
        r = requests.get(url+"projects?name="+project+"", auth=("admin", "Harbor12345"), verify=False)
        projectid = r.json()[0]['project_id']
        r = requests.get(url+"targets?name="+target+"", auth=("admin", "Harbor12345"), verify=False)
        targetid = r.json()[0]['id']
        payload = {"name": ""+rulename+"", "description": "string", "projects": [{"project_id": projectid,}], "targets": [{"id": targetid,}], "trigger": {"kind": ""+trigger+"", "schedule_param": {"type": "weekly", "weekday": 1, "offtime": 0}}}
        r = requests.post(url+"policies/replication", auth=("admin", "Harbor12345"), json=payload, verify=False)
        print(r.status_code)
    
    def updateprojectsetting(project, contenttrust, preventrunning, preventseverity, scanonpush):
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
    
    def updatesystemsetting(emailfrom, emailhost, emailport, emailuser, creation, selfreg, token):
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
    
    def updaterepoinfo(reponame):
        r = requests.put(url+"repositories/"+reponame+"", auth=("admin", "Harbor12345"), json={"description": "testdescription"}, verify=False)
        print(r.status_code)


with open("testdata.json") as f:
    data = json.load(f)

def pullimage(*image):
    for i in image:
        os.system("docker pull "+i)

def pushimage(image, project):
    os.system("docker tag "+image+" "+host+"/"+project+"/"+image)
    os.system("docker login "+host+" -u Admin"+" -p Harbor12345")
    os.system("docker push "+host+"/"+project+"/"+image)

def pushsigned(image, project, tag):
    os.system("export DOCKER_CONTENT_TRUST=1;export DOCKER_CONTENT_TRUST_SERVER=https://"+host+":4443")
    os.system("export NOTARY_ROOT_PASSPHARSE=Harbor12345;export NOTARY_TARGETS_PASSPHRASE=Harbor12345;export NOTARY_SNAPSHOT_PASSPHRASE=Harbor12345")
    os.system("export DOCKER_CONTENT_TRUST_ROOT_PASSPHRASE=Harbor12345; export DOCKER_CONTENT_TRUST_REPOSITORY_PASSPHRASE=Harbor12345")
    os.system("export DOCKER_CONTENT_TRUST_OFFLINE_PASSPHRASE=Harbor12345; export DOCKER_CONTENT_TRUST_TAGGING_PASSPHRASE=Harbor12345")
    os.system("docker tag "+image+":"+tag+" "+host+"/"+project+"/"+image+":"+tag)
    os.system("docker login "+host+" -u Admin"+" -p Harbor12345")
    os.system("docker push "+host+"/"+project+"/"+image+":"+tag)

def createonedotone():
    for user in data["users"]:
        Vonedotone.createuser(user["name"])
    for user in data["admin"]:
        Vonedotone.setuseradmin(user["name"])
    for project in data["projects"]:
        Vonedotone.createproject(project["name"])
        for member in project["member"]: 
            Vonedotone.addmember(project["name"], member["name"], member["role"])
    pullimage("busybox", "redis", "haproxy", "alpine", "httpd:2")
    pushimage("busybox", data["projects"][0]["name"])
    if protocol == "https":
        pushsigned("alpine", data["projects"][0]["name"], "latest")
    else:
        print("http does not support notary")

    for endpoint in data["endpoint"]:
        Vonedotone.addendpoint(endpoint["url"], endpoint["name"], endpoint["user"], endpoint["pass"], False)
    for replicationrule in data["replicationrule"]:
        Vonedotone.addreplicationrule(replicationrule["project"], replicationrule["endpoint"], replicationrule["rulename"], 0)
    Vonedotone.updateprojectsetting(data["projects"][0]["name"], 1)
    ef = data["configuration"]["emailsetting"]["emailfrom"]
    eh = data["configuration"]["emailsetting"]["emailserver"]
    ep = data["configuration"]["emailsetting"]["emailport"]
    eu = data["configuration"]["emailsetting"]["emailuser"]
    creation = data["configuration"]["projectcreation"]
    Vonedotone.updatesystemsetting(ef, eh, ep, eu, creation)

def createonedottwo():
    for user in data["users"]:
        Vonedottwo.createuser(user["name"])
    for user in data["admin"]:
        Vonedottwo.setuseradmin(user["name"])
    for project in data["projects"]:
        Vonedottwo.createproject(project["name"])
        for member in project["member"]:
            Vonedottwo.addmember(project["name"], member["name"], member["role"])
    pullimage("busybox", "redis", "haproxy", "alpine", "httpd:2")
    pushimage("busybox", data["projects"][0]["name"])
    if protocol == "https":
        pushsigned("alpine", data["projects"][0]["name"], "latest")
    else:
        print("http does not support notary")

    for endpoint in data["endpoint"]:
        Vonedottwo.addendpoint(endpoint["url"], endpoint["name"], endpoint["user"], endpoint["pass"], False)
    for replicationrule in data["replicationrule"]:
        Vonedottwo.addreplicationrule(replicationrule["project"], replicationrule["endpoint"], replicationrule["rulename"], 0)
    Vonedottwo.updateprojectsetting(data["projects"][0]["name"], 1)
    ef = data["configuration"]["emailsetting"]["emailfrom"]
    eh = data["configuration"]["emailsetting"]["emailserver"]
    ep = float(data["configuration"]["emailsetting"]["emailport"])
    eu = data["configuration"]["emailsetting"]["emailuser"]
    creation = data["configuration"]["projectcreation"]
    token = data["configuration"]["token"]
    selfreg = data["configuration"]["selfreg"]
    Vonedottwo.updatesystemsetting(ef, eh, ep, eu, creation, selfreg, token)

def createonedotthree():
    for user in data["users"]:
        Vonedotthree.createuser(user["name"])
    for user in data["admin"]:
        Vonedotthree.setuseradmin(user["name"])
    for project in data["projects"]:
        Vonedotthree.createproject(project["name"])
        for member in project["member"]: 
            Vonedotthree.addmember(project["name"], member["name"], member["role"])
    pullimage("busybox", "redis", "haproxy", "alpine", "httpd:2")
    pushimage("busybox", data["projects"][0]["name"])
    if protocol == "https":
        pushsigned("alpine", data["projects"][0]["name"], "latest")
    else:
        print("http does not support notary")

    for endpoint in data["endpoint"]:
        Vonedotthree.addendpoint(endpoint["url"], endpoint["name"], endpoint["user"], endpoint["pass"], False)
    for replicationrule in data["replicationrule"]:
        Vonedotthree.addreplicationrule(replicationrule["project"], replicationrule["endpoint"], replicationrule["rulename"], 0)
    for project in data["projects"]:
        ct = project["configuration"]["enable_content_trust"]
        pr = project["configuration"]["prevent_vulnerable_images_from_running"]
        prs = project["configuration"]["prevent_vlunerable_images_from_running_severity"]
        sop = project["configuration"]["automatically_scan_images_on_push"]
        print(ct, pr, prs, sop)
        Vonedotthree.updateprojectsetting(project["name"], ct, pr, prs, sop)
    ef = data["configuration"]["emailsetting"]["emailfrom"]
    eh = data["configuration"]["emailsetting"]["emailserver"]
    ep = float(data["configuration"]["emailsetting"]["emailport"])
    eu = data["configuration"]["emailsetting"]["emailuser"]
    creation = data["configuration"]["projectcreation"]
    token = data["configuration"]["token"]
    selfreg = data["configuration"]["selfreg"]
    Vonedotthree.updatesystemsetting(ef, eh, ep, eu, creation, selfreg, token)

def createonedotfour():
    for user in data["users"]:
        Vonedotfour.createuser(user["name"])
    for user in data["admin"]:
        Vonedotfour.setuseradmin(user["name"])
    for project in data["projects"]:
        Vonedotfour.createproject(project["name"])
        for member in project["member"]: 
            Vonedotfour.addmember(project["name"], member["name"], member["role"])
    pullimage("busybox", "redis", "haproxy", "alpine", "httpd:2")
    pushimage("busybox", data["projects"][0]["name"])
    if protocol == "https":
        pushsigned("alpine", data["projects"][0]["name"], "latest")
    else:
        print("http does not support notary")

    for endpoint in data["endpoint"]:
        Vonedotfour.addendpoint(endpoint["url"], endpoint["name"], endpoint["user"], endpoint["pass"], False)
    for replicationrule in data["replicationrule"]:
        Vonedotfour.addreplicationrule(replicationrule["project"], replicationrule["endpoint"], replicationrule["trigger"], replicationrule["rulename"])
    for project in data["projects"]:
        ct = project["configuration"]["enable_content_trust"]
        pr = project["configuration"]["prevent_vulnerable_images_from_running"]
        prs = project["configuration"]["prevent_vlunerable_images_from_running_severity"]
        sop = project["configuration"]["automatically_scan_images_on_push"]
        Vonedotfour.updateprojectsetting(project["name"], ct, pr, prs, sop)
    ef = data["configuration"]["emailsetting"]["emailfrom"]
    eh = data["configuration"]["emailsetting"]["emailserver"]
    ep = float(data["configuration"]["emailsetting"]["emailport"])
    eu = data["configuration"]["emailsetting"]["emailuser"]
    creation = data["configuration"]["projectcreation"]
    token = data["configuration"]["token"]
    selfreg = data["configuration"]["selfreg"]
    Vonedotfour.updatesystemsetting(ef, eh, ep, eu, creation, selfreg, token)


def createonedotfive():
    for user in data["users"]:
        Vonedotfive.createuser(user["name"])
    for user in data["admin"]:
        Vonedotfive.setuseradmin(user["name"])
    for project in data["projects"]:
        Vonedotfive.createproject(project["name"])
        for member in project["member"]: 
            Vonedotfive.addmember(project["name"], member["name"], member["role"])
        for label in project["labels"]:
            Vonedotfive.addprojectlabel(project["name"], label["name"])
    for label in data["configuration"]["syslabel"]:
        Vonedotfive.addsyslabel(label["name"])
    pullimage("busybox", "redis", "haproxy", "alpine", "httpd:2")
    pushimage("busybox", data["projects"][0]["name"])
    if protocol == "https":
        pushsigned("alpine", data["projects"][0]["name"], "latest")
    else:
        print("http does not support notary")

    for endpoint in data["endpoint"]:
        Vonedotfive.addendpoint(endpoint["url"], endpoint["name"], endpoint["user"], endpoint["pass"], False)
    for replicationrule in data["replicationrule"]:
        Vonedotfive.addreplicationrule(replicationrule["project"], replicationrule["endpoint"], replicationrule["trigger"], replicationrule["rulename"])
    for project in data["projects"]:
        ct = project["configuration"]["enable_content_trust"]
        pr = project["configuration"]["prevent_vulnerable_images_from_running"]
        prs = project["configuration"]["prevent_vlunerable_images_from_running_severity"]
        sop = project["configuration"]["automatically_scan_images_on_push"]
        Vonedotfive.updateprojectsetting(project["name"], ct, pr, prs, sop)
    ef = data["configuration"]["emailsetting"]["emailfrom"]
    eh = data["configuration"]["emailsetting"]["emailserver"]
    ep = float(data["configuration"]["emailsetting"]["emailport"])
    eu = data["configuration"]["emailsetting"]["emailuser"]
    creation = data["configuration"]["projectcreation"]
    token = data["configuration"]["token"]
    selfreg = data["configuration"]["selfreg"]
    Vonedotfive.updatesystemsetting(ef, eh, ep, eu, creation, selfreg, token)

if version == 1.1:
    createdata = Vonedotone()
    createonedotone()
elif version == 1.2:
    createdata = Vonedottwo()
    createonedottwo()
elif version == 1.3:
    createdata = Vonedotthree()
    createonedotthree()
elif version == 1.4:
    createdata = Vonedotfour()  
    createonedotfour()
elif version == 1.5:
    createdata = Vonedotfive()
    createonedotfive()
else:  
    print("version not supported")



