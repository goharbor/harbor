#!/usr/bin/env python

import json
import logging
import requests

logging.basicConfig(level=logging.INFO)


class HarborClient(object):
    def __init__(self, host, user, password, protocol="http"):
        self.host = host
        self.user = user
        self.password = password
        self.protocol = protocol

        self.session_id = self.login()

    def __del__(self):
        self.logout()

    def login(self):
        login_data = requests.post('%s://%s/login' %
                                   (self.protocol, self.host),
                                   data={'principal': self.user,
                                         'password': self.password})
        if login_data.status_code == 200:
            session_id = login_data.cookies.get('beegosessionID')

            logging.debug("Successfully login, session id: {}".format(
                session_id))
            return session_id
        else:
            logging.error("Fail to login, please try again")
            return None

    def logout(self):
        requests.get('%s://%s/logout' % (self.protocol, self.host),
                     cookies={'beegosessionID': self.session_id})
        logging.debug("Successfully logout")

    # Get project id
    def get_project_id_from_name(self, project_name):
        registry_data = requests.get(
            '%s://%s/api/projects?project_name=%s' %
            (self.protocol, self.host, project_name),
            cookies={'beegosessionID': self.session_id})
        if registry_data.status_code == 200 and registry_data.json():
            project_id = registry_data.json()[0]['project_id']
            logging.debug(
                "Successfully get project id: {}, project name: {}".format(
                    project_id, project_name))
            return project_id
        else:
            logging.error("Fail to get project id from project name",
                          project_name)
            return None

    # GET /search
    def search(self, query_string):
        result = None
        path = '%s://%s/api/search?q=%s' % (self.protocol, self.host,
                                            query_string)
        response = requests.get(path,
                                cookies={'beegosessionID': self.session_id})
        if response.status_code == 200:
            result = response.json()
            logging.debug("Successfully get search result: {}".format(result))
        else:
            logging.error("Fail to get search result")
        return result

    # GET /projects
    def get_projects(self, project_name=None, is_public=None):
        # TODO: support parameter
        result = None
        path = '%s://%s/api/projects' % (self.protocol, self.host)
        response = requests.get(path,
                                cookies={'beegosessionID': self.session_id})
        if response.status_code == 200:
            result = response.json()
            logging.debug("Successfully get projects result: {}".format(
                result))
        else:
            logging.error("Fail to get projects result")
        return result

    # HEAD /projects
    def check_project_exist(self, project_name):
        result = False
        path = '%s://%s/api/projects?project_name=%s' % (
            self.protocol, self.host, project_name)
        response = requests.head(path,
                                 cookies={'beegosessionID': self.session_id})
        if response.status_code == 200:
            result = True
            logging.debug(
                "Successfully check project exist, result: {}".format(result))
        elif response.status_code == 404:
            result = False
            logging.debug(
                "Successfully check project exist, result: {}".format(result))
        else:
            logging.error("Fail to check project exist")
        return result

    # POST /projects
    def create_project(self, project_name, is_public=False):
        result = False
        path = '%s://%s/api/projects' % (self.protocol, self.host)
        request_body = json.dumps({'project_name': project_name,
                                   'public': is_public})
        response = requests.post(path,
                                 cookies={'beegosessionID': self.session_id},
                                 data=request_body)
        if response.status_code == 201 or response.status_code == 500:
            # TODO: the response return 500 sometimes
            result = True
            logging.debug(
                "Successfully create project with project name: {}".format(
                    project_name))
        else:
            logging.error(
                "Fail to create project with project name: {}, response code: {}".format(
                    project_name, response.status_code))
        return result
   
   # DELETE /projects/{project_id}
    def delete_project(self, project_id):
        result = False
        path = '%s://%s/api/projects?project_id=%s' % (self.protocol, self.host, project_id)
        response = requests.delete(path,
                                 cookies={'beegosessionID': self.session_id})
        if response.status_code == 200:
            result = True
            logging.debug(
                "Successfully delete project with project id: {}".format(
                    project_id))
        else:
            logging.error(
                "Fail to delete project with project id: {}, response code: {}".format(
                    project_id, response.status_code))
        return result

   # GET /projects/{project_id}
    def get_project_info(self, project_id):
        result = None
        path = '%s://%s/api/projects?project_id=%s' % (self.protocol, self.host, project_id)
        response = requests.get(path,
                                cookies={'beegosessionID': self.session_id})
        if response.status_code == 200:
            result = response.json()
            logging.debug("Successfully get project info: {}".format(
                result))
        else:
            logging.error("Fail to get project info")
        return result

    # PUT /projects/{project_id}/publicity
    def set_project_publicity(self, project_id, is_public):
        result = False
        path = '%s://%s/api/projects/%s/publicity?project_id=%s' % (
            self.protocol, self.host, project_id, project_id)
        request_body = json.dumps({'public': is_public})
        response = requests.put(path,
                                cookies={'beegosessionID': self.session_id},
                                data=request_body)
        if response.status_code == 200:
            result = True
            logging.debug(
                "Success to set project id: {} with publicity: {}".format(
                    project_id, is_public))
        else:
            logging.error(
                "Fail to set publicity to project id: {} with status code: {}".format(
                    project_id, response.status_code))
        return result

    # POST /projects/{project_id}/logs/filter
    def get_project_logs(self, project_id, page, page_size, username, keywords, begin_time, end_time):
        result = None
        path = '%s://%s/api/projects/%s/logs/filter?page=%s&page_size=%s' % (self.protocol, self.host, project_id, page, page_size)
        request_body = json.dumps({'username': username,
                                   'keywords': keywords,
                                   'begin_timestamp': begin_time,
                                   'end_timestamp': end_time})
        response = requests.post(path,
                                 cookies={'beegosessionID': self.session_id},
                                 data=request_body)
        if response.status_code == 200:
            result = response.json()
            logging.debug(
                "Successfully get project log: {}".format(result))
        else:
            logging.error("Fail to get project log")
        return result

    # GET /projects/{project_id}/members
    def get_project_members(self, project_id):
        result=None
        path='%s://%s/api/projects/%s/members' % (self.protocol, self.host, project_id)
        response = requests.get(path, cookies={'beegosessionID': self.session_id})
        if response.status_code == 200:
            result =response.json()
            logging.debug("Successfully get project members: {}".format(result))
        else:
            logging.error("Fail to get project members")

    # POST /projects/{project_id}/memebers
    def add_user_project_role(self, project_id, role, user_name):
        result = False
        path = '%s://%s/api/projects/%s/members' % (self.protocol, self.host, project_id)
        request_body = json.dumps({'roles': role,
                                   'username': user_name})
        response = requests.post(path,
                                 cookies={'beegosessionID': self.session_id},
                                 data=request_body)
        if response.status_code == 200:
            result = True
            logging.debug(
                "Successfully add user project role: {}")
        else:
            logging.error(
                "Fail to add user project role, response code: {}".format(response.status_code))
        return result

    # GET /projects/{project_id}/members/{user_id}
    def get_user_project_role(self, project_id, user_id):
        result = None
        path = '%s://%s/api/projects/%s/members/%s' % (self.protocol, self.host, project_id, user_id)
        response = requests.get(path,
                                cookies={'beegosessionID': self.session_id})
        if response.status_code == 200:
            result = response.json()
            logging.debug("Successfully get user project role: {}".format(
                result))
        else:
            logging.error("Fail to get user project role.")
        return result

    # PUT /projects/{project_id}/members/{user_id}
    def update_user_project_role(self, project_id, user_id, role, user_name):
        result = False
        path = '%s://%s/api/projects/%s/members/%s' % (
            self.protocol, self.host, project_id, user_id)
        request_body = json.dumps({'roles': role,
                                    'username': user_name})
        response = requests.put(path,
                                cookies={'beegosessionID': self.session_id},
                                data=request_body)
        if response.status_code == 200:
            result = True
            logging.debug(
                "Successfully update user project role.")
        else:
            logging.error(
                "Fail to update user project role with status code: {}".format(response.status_code))
        return result

    # DELETE /projects/{project_id}/members/{user_id}
    def delete_user_project_role(self, project_id, user_id):
        result = False
        path = '%s://%s/api/projects/%s/members/%s' % (self.protocol, self.host, project_id, user_id)
        response = requests.delete(path,
                                 cookies={'beegosessionID': self.session_id})
        if response.status_code == 200:
            result = True
            logging.debug(
                "Successfully delete user project role")
        else:
            logging.error(
                "Fail to delete user project role, response code: {}".format(
                    response.status_code))
        return result

    # GET /statistics
    def get_statistics(self):
        result = None
        path = '%s://%s/api/statistics' % (self.protocol, self.host)
        response = requests.get(path,
                                cookies={'beegosessionID': self.session_id})
        if response.status_code == 200:
            result = response.json()
            logging.debug("Successfully get statistics: {}".format(result))
        else:
            logging.error("Fail to get statistics result")
        return result

    # GET /users
    def get_users(self, user_name):
        # TODO: support parameter
        result = None
        path = '%s://%s/api/users?/username=%s' % (self.protocol, self.host, user_name)
        response = requests.get(path,
                                cookies={'beegosessionID': self.session_id})
        if response.status_code == 200:
            result = response.json()
            logging.debug("Successfully get users result: {}".format(result))
        else:
            logging.error("Fail to get users result")
        return result

    # POST /users
    def create_user(self, username, email, password, realname, comment):
        result = False
        path = '%s://%s/api/users' % (self.protocol, self.host)
        request_body = json.dumps({'username': username,
                                   'email': email,
                                   'password': password,
                                   'realname': realname,
                                   'comment': comment})
        response = requests.post(path,
                                 cookies={'beegosessionID': self.session_id},
                                 data=request_body)
        if response.status_code == 201:
            result = True
            logging.debug("Successfully create user with username: {}".format(
                username))
        else:
            logging.error(
                "Fail to create user with username: {}, response code: {}".format(
                    username, response.status_code))
        return result

    # GET /users
    def get_current_user(self):
        result = None
        path = '%s://%s/api/users/current' % (self.protocol, self.host)
        response = requests.get(path,
                                cookies={'beegosessionID': self.session_id})
        if response.status_code == 200:
            result = response.json()
            logging.debug("Successfully get current user: {}".format(result))
        else:
            logging.error("Fail to get current user")
        return result

    # PUT /users/{user_id}
    def update_user_profile(self, user_id, email, realname, comment):
        # TODO: support not passing comment
        result = False
        path = '%s://%s/api/users/%s' % (self.protocol, self.host,
                                                    user_id)
        request_body = json.dumps({'email': email,
                                   'realname': realname,
                                   'comment': comment})
        response = requests.put(path,
                                cookies={'beegosessionID': self.session_id},
                                data=request_body)
        if response.status_code == 200:
            result = True
            logging.debug(
                "Successfully update user profile with user id: {}".format(
                    user_id))
        else:
            logging.error(
                "Fail to update user profile with user id: {}, response code: {}".format(
                    user_id, response.status_code))
        return result

    # DELETE /users/{user_id}
    def delete_user(self, user_id):
        result = False
        path = '%s://%s/api/users/%s' % (self.protocol, self.host,
                                                    user_id)
        response = requests.delete(path,
                                   cookies={'beegosessionID': self.session_id})
        if response.status_code == 200:
            result = True
            logging.debug("Successfully delete user with id: {}".format(
                user_id))
        else:
            logging.error("Fail to delete user with id: {}".format(user_id))
        return result

    # PUT /users/{user_id}/password
    def change_password(self, user_id, old_password, new_password):
        result = False
        path = '%s://%s/api/users/%s/password' % (
            self.protocol, self.host, user_id)
        request_body = json.dumps({'old_password': old_password,
                                   'new_password': new_password})
        response = requests.put(path,
                                cookies={'beegosessionID': self.session_id},
                                data=request_body)
        if response.status_code == 200:
            result = True
            logging.debug(
                "Successfully change password for user id: {}".format(user_id))
        else:
            logging.error("Fail to change password for user id: {}".format(
                user_id))
        return result

    # PUT /users/{user_id}/sysadmin
    def toggle_admin(self, user_id, has_admin_role):
        result = False
        path = '%s://%s/api/users/%s/sysadmin' % (
            self.protocol, self.host, user_id)
        request_body = json.dumps({'has_admin_role': has_admin_role})
        response = requests.put(path,
                                cookies={'beegosessionID': self.session_id})
        if response.status_code == 200:
            result = True
            logging.debug(
                "Successfully toggle user admin with user id: {}".format(
                    user_id))
        else:
            logging.error(
                "Fail to toggle user admin with user id: {}, response code: {}".format(
                    user_id, response.status_code))
        return result

    # GET /repositories
    def get_repositories(self, project_id, detail, repo_name, page, page_size):
        result = None
        path = '%s://%s/api/repositories?project_id=%s&detail=%s&q=%s&page=%s&page_size=%s' % (
            self.protocol, self.host, project_id, detail, repo_name, page, page_size)
        response = requests.get(path,
                                cookies={'beegosessionID': self.session_id})
        if response.status_code == 200:
            result = response.json()
            logging.debug(
                "Successfully get repositories with id: {}, result: {}".format(
                    project_id, result))
        else:
            logging.error("Fail to get repositories result with id: {}".format(
                project_id))
        return result

    # DELETE /repositories{repo_name}/tags/{tag}
    def delete_repository_tag(self, repo_name, tag):
        # TODO: support to check tag
        # TODO: return 200 but the repo is not deleted, need more test
        result = False
        path = '%s://%s/api/repositories/%s/tags/%s' % (self.protocol,
                                                          self.host, repo_name, tag)
        response = requests.delete(path,
                                   cookies={'beegosessionID': self.session_id})
        if response.status_code == 200:
            result = True
            logging.debug("Successfully delete repository {} with tag {}".format(
                repo_name, tag))
        else:
            logging.error("Fail to delete repository {} with tag {}".format(repo_name, tag))
        return result

    # Get /repositories/{repo_name}/tags
    def get_repository_tags(self, repo_name, detail):
        result = None
        path = '%s://%s/api/repositories/%s/tags?detail=%s' % (
            self.protocol, self.host, repo_name, detail)
        response = requests.get(path,
                                cookies={'beegosessionID': self.session_id})
        if response.status_code == 200:
            result = response.json()
            logging.debug(
                "Successfully get tags with repo name: {}, result: {}".format(
                    repo_name, result))
        else:
            logging.error("Fail to get tags with repo name: {}".format(
                repo_name))
        return result

    # DELETE /repositories{repo_name}/tags
    def delete_repository_tags(self, repo_name):
        # TODO: return 200 but the repo is not deleted, need more test
        result = False
        path = '%s://%s/api/repositories/%s/tags' % (self.protocol,
                                                          self.host, repo_name)
        response = requests.delete(path,
                                   cookies={'beegosessionID': self.session_id})
        if response.status_code == 200:
            result = True
            logging.debug("Successfully delete all tags of repository {}".format(
                repo_name))
        else:
            logging.error("Fail to delete all tags of repository {}".format(repo_name))
        return result

    # GET /repositories/{repo_name}/tags/{tag}/manifest
    def get_repository_manifest(self, repo_name, tag, version):
        result = None
        path = '%s://%s/api/repositories/%s/tags/%s/manifest?version=%s' % (
            self.protocol, self.host, repo_name, tag, version)
        response = requests.get(path,
                                cookies={'beegosessionID': self.session_id})
        if response.status_code == 200:
            result = response.json()
            logging.debug(
                "Successfully get manifest with repo name: {}, tag: {}, result: {}".format(
                    repo_name, tag, result))
        else:
            logging.error(
                "Fail to get manifest with repo name: {}, tag: {}".format(
                    repo_name, tag))
        return result

    # GET /repositories/{repo_name}/signatures
    def get_repository_signatures(self, repo_name):
        result = None
        path = '%s://%s/api/repositories/%s/signatures' % (
            self.protocol, self.host, repo_name)
        response = requests.get(path,
                                cookies={'beegosessionID': self.session_id})
        if response.status_code == 200:
            result = response.json()
            logging.debug(
                "Successfully get signatures with repo name: {}, result: {}".format(
                    repo_name, result))
        else:
            logging.error(
                "Fail to get signatures with repo name: {}".format(
                    repo_name))
        return result

    # GET /repositories/top
    def get_top_accessed_repositories(self, count=None, detail=None):
        result = None
        path = '%s://%s/api/repositories/top' % (self.protocol, self.host)
        if count:
            path += "?count=%s" % (count)
            if detail:
                path += "&detail=%s" % (detail)
        else:
            if detail:
                path += "?detail=%s" % (detail)
        response = requests.get(path,
                                cookies={'beegosessionID': self.session_id})
        if response.status_code == 200:
            result = response.json()
            logging.debug(
                "Successfully get top accessed repositories, result: {}".format(
                    result))
        else:
            logging.error("Fail to get top accessed repositories")
        return result

    # GET /logs
    def get_logs(self, lines=None, start_time=None, end_time=None):
        result = None
        path = '%s://%s/api/logs' % (self.protocol, self.host)
        response = requests.get(path,
                                cookies={'beegosessionID': self.session_id})
        if response.status_code == 200:
            result = response.json()
            logging.debug("Successfully get logs")
        else:
            logging.error("Fail to get logs and response code: {}".format(
                response.status_code))
        return result

    # GET /jobs/replication
    def get_jobs(self, policy_id=None, num=None, start_time=None, end_time=None, repository=None, status=None, page=None, page_size=None):
        #TODO add optional parameters
        result = None
        path = '%s://%s/api/jobs/replication' % (self.protocol, self.host)
        if policy_id:
            path += "?policy_id=%s" % (policy_id)
        response = requests.get(path,
                                cookies={'beegosessionID': self.session_id})
        if response.status_code == 200:
            result = response.json()
            logging.debug("Successfully get jobs")
        else:
            logging.error("Fail to get jobs and response code: {}".format(
                response.status_code))
        return result

     # DELETE /jobs/replication/{id}
    def delete_job(self, job_id):
        result = False
        path = '%s://%s/api/jobs/replication/%s' % (self.protocol,
                                                          self.host, job_id)
        response = requests.delete(path,
                                   cookies={'beegosessionID': self.session_id})
        if response.status_code == 200:
            result = True
            logging.debug("Successfully delete job with id: {}".format(
                job_id))
        else:
            logging.error("Fail to delete job with id: {} and response code {}".format(job_id, response.status_code))
        return result

    # GET /jobs/replication/{id}/log
    def get_job_logs(self, job_id):
        result = None
        path = '%s://%s/api/jobs/replication/%s/log' % (self.protocol, self.host, job_id)
        response = requests.get(path,
                                cookies={'beegosessionID': self.session_id})
        if response.status_code == 200:
            result = response.json()
            logging.debug("Successfully get job logs")
        else:
            logging.error("Fail to get job logs and response code: {}".format(
                response.status_code))
        return result

    # GET /policies/replication
    def get_policies(self, name=None, project_id=None):
        result = None
        path = '%s://%s/api/policies/replication' % (self.protocol, self.host)
        if name:
            path += "?name=%s" % (name)
            if detail:
                path += "&project_id=%s" % (project_id)
        else:
            if project_id:
                path += "?project_id=%s" % (project_id)
        response = requests.get(path,
                                cookies={'beegosessionID': self.session_id})
        if response.status_code == 200:
            result = response.json()
            logging.debug(
                "Successfully get policies, result: {}".format(
                    result))
        else:
            logging.error("Fail to get policies with response code {}".format(response.status_code))
        return result

    # POST /policies/replication
    def create_policy(self, project_id, target_id, name):
        result = False
        path = '%s://%s/api/policies/replication' % (self.protocol, self.host)
        request_body = json.dumps({'project_id': project_id,
                                   'target_id': target_id,
                                   'name': name})
        response = requests.post(path,
                                 cookies={'beegosessionID': self.session_id},
                                 data=request_body)
        if response.status_code == 201:
            result = True
            logging.debug("Successfully create policy with name: {}".format(
                name))
        else:
            logging.error(
                "Fail to create user with name: {}, response code: {}".format(
                    name, response.status_code))
        return result

    # GET /policies/replication/{id}
    def get_policy(self, policy_id):
        result = None
        path = '%s://%s/api/policies/replication/%s' % (self.protocol, self.host, policy_id)
        response = requests.get(path,
                                cookies={'beegosessionID': self.session_id})
        if response.status_code == 200:
            result = response.json()
            logging.debug("Successfully get policy info")
        else:
            logging.error("Fail to get policy info and response code: {}".format(
                response.status_code))
        return result

    # PUT /policies/replication/{id}
    def update_policy(self, policy_id, target_id, name, enabled, description, cron_str):
        result = False
        path = '%s://%s/api/policies/replication/%s' % (
            self.protocol, self.host, policy_id)
        request_body = json.dumps({'target_id': target_id,
                                   'name': name,
                                   'enabled': enabled,
                                   'description': description,
                                   'cron_str': cron_str})
        response = requests.put(path,
                                cookies={'beegosessionID': self.session_id},
                                data=request_body)
        if response.status_code == 200:
            result = True
            logging.debug(
                "Successfully update policy for policy id: {}".format(policy_id))
        else:
            logging.error("Fail to update policy for policy id: {} with response code {}".format(policy_id, response.status_code))
        return result

    # PUT /policies/replication/{id}/enablement
    def update_policy_enablement(self, policy_id, enabled):
        result = False
        path = '%s://%s/api/policies/replication/%s/enablement' % (
            self.protocol, self.host, policy_id)
        request_body = json.dumps({'enabled': enabled})
        response = requests.put(path,
                                cookies={'beegosessionID': self.session_id},
                                data=request_body)
        if response.status_code == 200:
            result = True
            logging.debug(
                "Successfully update policy enablement for policy id: {}".format(policy_id))
        else:
            logging.error("Fail to update policy enablement for policy id: {} with response code".format(policy_id, response.status_code))
        return result

    # GET /targets
    def get_replication_targets(self, name=None):
        result = None
        path = '%s://%s/api/targets' % (self.protocol, self.host)
        if name:
            path += "?name=%s" % (name)
        response = requests.get(path,
                                cookies={'beegosessionID': self.session_id})
        if response.status_code == 200:
            result = response.json()
            logging.debug("Successfully get targets")
        else:
            logging.error("Fail to get targets and response code: {}".format(
                response.status_code))
        return result

    # POST /targets
    def create_replication_target(self, endpoint, name, user_name, password):
        result = False
        path = '%s://%s/api/targets' % (self.protocol, self.host)
        request_body = json.dumps({'endpoint': endpoint,
                                   'name': name,
                                   'username': user_name,
                                   'password': password})
        response = requests.post(path,
                                 cookies={'beegosessionID': self.session_id},
                                 data=request_body)
        if response.status_code == 201:
            result = True
            logging.debug("Successfully create replication target with name: {}".format(
                name))
        else:
            logging.error(
                "Fail to create replication target with name: {}, response code: {}".format(
                    name, response.status_code))
        return result

    # POST /targets/ping
    def ping_replication_target(self, endpoint, user_name, password):
        result = False
        path = '%s://%s/api/targets/ping' % (self.protocol, self.host)
        request_body = json.dumps({'endpoint': endpoint,
                                   'username': user_name,
                                   'password': password})
        response = requests.post(path,
                                 cookies={'beegosessionID': self.session_id},
                                 data=request_body)
        if response.status_code == 200:
            result = True
            logging.debug("Successfully ping replication target ")
        else:
            logging.error(
                "Fail to ping replication target, response code: {}".format(
                    response.status_code))
        return result

    # POST /targets/{id}/ping
    def ping_replication_target_with_id(self, target_id):
        result = False
        path = '%s://%s/api/targets/%s/ping' % (self.protocol, self.host, target_id)
        response = requests.post(path,
                                 cookies={'beegosessionID': self.session_id},
                                 data=request_body)
        if response.status_code == 200:
            result = True
            logging.debug("Successfully ping replication target with id: {} ".format(target_id))
        else:
            logging.error(
                    "Fail to ping replication target with id: {}, response code: {}".format(target_id,
                    response.status_code))
        return result

    # POST /targets/{id}
    def update_replication_target(self,  target_id, endpoint, name, user_name, password):
        result = False
        path = '%s://%s/api/targets/%s' % (self.protocol, self.host, target_id)
        request_body = json.dumps({'endpoint': endpoint,
                                   'name': name,
                                   'username': user_name,
                                   'password': password})
        response = requests.put(path,
                                 cookies={'beegosessionID': self.session_id},
                                 data=request_body)
        if response.status_code == 200:
            result = True
            logging.debug("Successfully update replication target with name: {}".format(
                name))
        else:
            logging.error(
                "Fail to update replication target with name: {}, response code: {}".format(
                    name, response.status_code))
        return result

    # GET /targets/{id}
    def get_replication_target(self, target_id):
        result = None
        path = '%s://%s/api/targets/%s' % (self.protocol, self.host, target_id)
        response = requests.get(path,
                                cookies={'beegosessionID': self.session_id})
        if response.status_code == 200:
            result = response.json()
            logging.debug("Successfully get target")
        else:
            logging.error("Fail to get target and response code: {}".format(
                response.status_code))
        return result

    # GET /targets/{id}
    def delete_replication_target(self, target_id):
        result = None
        path = '%s://%s/api/targets/%s' % (self.protocol, self.host, target_id)
        response = requests.delete(path,
                                cookies={'beegosessionID': self.session_id})
        if response.status_code == 200:
            result = response.json()
            logging.debug("Successfully delete replication target")
        else:
            logging.error("Fail to delete replication target and response code: {}".format(
                response.status_code))
        return result

    # GET /targets/{id}/policies
    def get_replication_target_policies(self, target_id):
        result = None
        path = '%s://%s/api/targets/%s/policies' % (self.protocol, self.host, target_id)
        response = requests.get(path,
                                cookies={'beegosessionID': self.session_id})
        if response.status_code == 200:
            result = response.json()
            logging.debug("Successfully get target policies")
        else:
            logging.error("Fail to get target policies and response code: {}".format(
                response.status_code))
        return result


