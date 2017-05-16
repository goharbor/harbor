#!/usr/bin/env python

import json
import logging
import requests

logging.basicConfig(level=logging.WARNING)

class HarborClient(object):
    def __init__(self, host, user, password, protocol="http"):
        self.host = host
        self.user = user
        self.password = password
        self.protocol = protocol

    def __del__(self):
        self.logout()

    def login(self):
        login_data = requests.post('%s://%s/login' %(self.protocol, self.host),
                                   data={'principal': self.user,
                                         'password': self.password}, verify=False)
        if login_data.status_code == 200:
            session_id = login_data.cookies.get('beegosessionID')
            self.session_id = session_id
            logging.debug("Successfully login, session id: {}".format(
                session_id))
        else:
            logging.error("Fail to login, please try again")

    def logout(self):
        requests.get('%s://%s/log_out' % (self.protocol, self.host),
                     cookies={'beegosessionID': self.session_id}, verify=False)
        logging.debug("Successfully logout")

    # GET /search
    def search(self, query_string):
        result = None
        path = '%s://%s/api/search?q=%s' % (self.protocol, self.host,
                                            query_string)
        response = requests.get(path,
                                cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = response.json()
            logging.debug("Successfully get search result: {}".format(result))
        else:
            logging.error("Fail to get search result")
        return result

    # GET /projects
    def get_projects(self):
        result = None
        path = '%s://%s/api/projects' % (self.protocol, self.host)
        response = requests.get(path,
                                cookies={'beegosessionID': self.session_id}, verify=False)
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
                                 cookies={'beegosessionID': self.session_id}, verify=False)
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
    def create_project(self, project_name, is_public=0):
        result = False
        path = '%s://%s/api/projects' % (self.protocol, self.host)
        request_body = json.dumps({'project_name': project_name,
                                   'public': is_public})
        response = requests.post(path,
                                 cookies={'beegosessionID': self.session_id},
                                 data=request_body, verify=False)
        if response.status_code == 201:
            result = True
            logging.debug(
                "Successfully create project with project name: {}".format(
                    project_name))
        else:
            logging.error(
                "Fail to create project with project name: {}, response code: {}".format(
                    project_name, response.status_code))
        return result

    # GET /projects/{project_id}/members
    def get_project_members(self, project_id):
        result = None
        path = '%s://%s/api/projects/%s/members' % (self.protocol, self.host, project_id)
        response = requests.get(path,
                                cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = response.json()
            logging.debug(
                "Successfully create project with project id: {}".format(
                    project_id))
        else:
            logging.error(
                "Fail to create project with project id: {}, response code: {}".format(
                    project_id, response.status_code))
        return result

    # POST /projects/{project_id}/members
    def add_project_member(self, project_id, username, role_id):
        result = False
        path = '%s://%s/api/projects/%s/members' % (self.protocol, self.host, project_id)
        request_str = '{"username": "%s","roles": [%s]}' % (username, role_id)
        request_body = json.dumps(json.loads(request_str))
        response = requests.post(path,
                                 cookies={'beegosessionID': self.session_id},
                                 data=request_body, verify=False)
        if response.status_code == 200:
            result = True
            logging.debug(
                "Successfully add project member with project id: {}".format(
                    project_id))
        else:
            logging.error(
                "Fail to add project member with project id: {}, response code: {}".format(
                    project_id, response.status_code))
        return result

    # DELETE /projects/{project_id}/members/{user_id}
    def delete_member_from_project(self, project_id, user_id):
        result = False
        path = '%s://%s/api/projects/%s/members/%s' % (self.protocol, self.host,
                                                       project_id, user_id)
        response = requests.delete(path,
                                   cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = True
            logging.debug("Successfully delete member with id: {}".format(
                user_id))
        else:
            logging.error("Fail to delete member with id: {}, response code: {}"
            .format(user_id, response.status_code))
        return result

    # PUT /projects/{project_id}/publicity
    def set_project_publicity(self, project_id, is_public):
        result = False
        path = '%s://%s/api/projects/%s/publicity' % (
            self.protocol, self.host, project_id)
        request_body = json.dumps({'public': is_public})
        response = requests.put(path,
                                cookies={'beegosessionID': self.session_id},
                                data=request_body, verify=False)
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

    # GET /statistics
    def get_statistics(self):
        result = None
        path = '%s://%s/api/statistics' % (self.protocol, self.host)
        response = requests.get(path,
                                cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = response.json()
            logging.debug("Successfully get statistics: {}".format(result))
        else:
            logging.error("Fail to get statistics result with status code: {}"
            .format(response.status_code))
        return result

    # GET /users
    def get_users(self):
        # TODO: support parameter
        result = None
        path = '%s://%s/api/users' % (self.protocol, self.host)
        response = requests.get(path,
                                cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = response.json()
            logging.debug("Successfully get users result: {}".format(result))
        else:
            logging.error("Fail to get users result with status code: {}"
            .format(response.status_code))
        return result

    # GET /users/current
    def get_user_info(self):
        result = None
        path = '%s://%s/api/users/current' % (self.protocol, self.host)
        response = requests.get(path,
                                cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = response.json()
            logging.debug("Successfully get users result: {}".format(result))
        else:
            logging.error("Fail to get users result with status code: {}"
            .format(response.status_code))
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
                                 data=request_body, verify=False)
        if response.status_code == 201:
            result = True
            logging.debug("Successfully create user with username: {}".format(
                username))
        else:
            logging.error(
                "Fail to create user with username: {}, response code: {}".format(
                    username, response.status_code))
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
                                data=request_body, verify=False)
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
                                   cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = True
            logging.debug("Successfully delete user with id: {}".format(
                user_id))
        else:
            logging.error("Fail to delete user with id: {}, response code: {}"
            .format(user_id, response.status_code))
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
                                data=request_body, verify=False)
        if response.status_code == 200:
            result = True
            logging.debug(
                "Successfully change password for user id: {}".format(user_id))
        else:
            logging.error("Fail to change password for user id: {}".format(
                user_id))
        return result

    # PUT /users/{user_id}/sysadmin
    def promote_as_admin(self, user_id, has_admin_role):
        result = False
        path = '%s://%s/api/users/%s/sysadmin' % (
            self.protocol, self.host, user_id)
        request_body = json.dumps({'has_admin_role': has_admin_role,
                                   'user_id': user_id})
        response = requests.put(path,
                                cookies={'beegosessionID': self.session_id},
                                data=request_body, verify=False)
        if response.status_code == 200:
            result = True
            logging.debug(
                "Successfully promote user as admin with user id: {}".format(
                    user_id))
        else:
            logging.error(
                "Fail to promote user as admin with user id: {}, response code: {}".format(
                    user_id, response.status_code))
        return result

    # GET /repositories
    def get_repositories(self, project_id, query_string=None):
        # TODO: support parameter
        result = None
        path = '%s://%s/api/repositories?project_id=%s' % (
            self.protocol, self.host, project_id)
        response = requests.get(path,
                                cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = response.json()
            logging.debug(
                "Successfully get repositories with id: {}, result: {}".format(
                    project_id, result))
        else:
            logging.error("Fail to get repositories result with id: {}, response code: {}".format(
                project_id, response.status_code))
        return result

    # DELETE /repositories/{repo_name}/tags/{tag}
    def delete_tag_of_repository(self, repo_name, tag):
        result = False
        path = '%s://%s/api/repositories/%s/tags/%s' % (self.protocol,self.host,
                                                        repo_name, tag)
        response = requests.delete(path,
                                   cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = True
            logging.debug("Successfully delete a tag of repository: {}".format(
                repo_name))
        else:
            logging.error("Fail to delete repository  with name: {}, response code: {}".format(
                repo_name, response.status_code))
        return result

    # DELETE /repositories/{repo_name}/tags
    def delete_tags_of_repository(self, repo_name):
        result = False
        path = '%s://%s/api/repositories/%s/tags' % (self.protocol,
                                                     self.host, repo_name)
        response = requests.delete(path,
                                   cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = True
            logging.debug("Successfully delete repository: {}".format(
                repo_name))
        else:
            logging.error("Fail to delete repository  with name: {}, response code: {}".format(
                repo_name, response.status_code))
        return result

    # Get /repositories/{repo_name}/tags
    def get_repository_tags(self, repo_name):
        result = None
        path = '%s://%s/api/repositories/%s/tags' % (
            self.protocol, self.host, repo_name)
        response = requests.get(path,
                                cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = response.json()
            logging.debug(
                "Successfully get tag with repo name: {}, result: {}".format(
                    repo_name, result))
        else:
            logging.error("Fail to get tags with repo name: {}, response code: {}".format(
                repo_name, response.status_code))
        return result

    # GET /repositories/{repo_name}/tags/{tag}/manifest
    def get_repository_manifest(self, repo_name, tag):
        result = None
        path = '%s://%s/api/repositories/%s/tags/%s/manifest' % (
            self.protocol, self.host, repo_name, tag)
        response = requests.get(path,
                                cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = response.json()
            logging.debug(
                "Successfully get manifests with repo name: {}, tag: {}, result: {}".format(
                    repo_name, tag, result))
        else:
            logging.error(
                "Fail to get manifests with repo name: {}, tag: {}".format(
                    repo_name, tag))
        return result

    # GET /repositories/top
    def get_top_accessed_repositories(self, count=None):
        result = None
        path = '%s://%s/api/repositories/top' % (self.protocol, self.host)
        if count:
            path += "?count=%s" % (count)
        response = requests.get(path,
                                cookies={'beegosessionID': self.session_id}, verify=False)
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
                                cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = response.json()
            logging.debug("Successfully get logs")
        else:
            logging.error("Fail to get logs and response code: {}".format(
                response.status_code))
        return result

    # Get /systeminfo
    def get_systeminfo(self):
        result = None
        path = '%s://%s/api/systeminfo' % (self.protocol, self.host)
        response = requests.get(path,
                                cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = response.json()
            logging.debug(
                "Successfully get systeminfo, result: {}".format(result))
        else:
            logging.error("Fail to get systeminfo, response code: {}".format(response.status_code))
        return result

    # Get /configurations
    def get_configurations(self):
        result = None
        path = '%s://%s/api/configurations' % (self.protocol, self.host)
        response = requests.get(path,
                                cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = response.json()
            logging.debug(
                "Successfully get configurations, result: {}".format(result))
        else:
            logging.error("Fail to get configurations, response code: {}".format(response.status_code))
        return result
