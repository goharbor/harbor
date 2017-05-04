#!/usr/bin/env python

import json
import logging
import requests

logging.basicConfig(level=logging.WARNING)


class HarborSdk(object):
    def __init__(self, host, user, password, protocol):
        self.host = host
        self.user = user
        self.password = password
        self.protocol = protocol

    def login(self):
        login_data = requests.post(
            '%s://%s/login' % (self.protocol, self.host),
            data={'principal': self.user,
                  'password': self.password},
            verify=False)
        if login_data.status_code == 200:
            session_id = login_data.cookies.get('beegosessionID')
            self.session_id = session_id
            logging.debug(
                "Successfully login, session id: {}".format(session_id))
        else:
            logging.error(
                "Fail to login, please try again, response code: {}, error: {}".
                    format(login_data.status_code, login_data.content))

    def logout(self):
        requests.get(
            '%s://%s/log_out' % (self.protocol, self.host),
            cookies={'beegosessionID': self.session_id},
            verify=False)
        logging.debug("Successfully logout")

    # GET /search
    def search(self, query_string):
        result = None
        path = '%s://%s/api/search?q=%s' % (self.protocol, self.host,
                                            query_string)
        response = requests.get(
            path, cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = response.json()
            logging.debug("Successfully get search result: {}".format(result))
        else:
            logging.error(
                "Fail to get search result, response code: {}, error: {}".
                    format(response.status_code, response.content))
        return result

    # GET /projects
    def get_projects(self,
                     name=None,
                     public=None,
                     owner=None,
                     page=None,
                     page_size=None):
        result = None
        path = '%s://%s/api/projects' % (self.protocol, self.host)
        payload = {'project_name': name,
                   'is_public': public,
                   'owner': owner,
                   'page': page,
                   'page_size': page_size}
        response = requests.get(
            path, params=payload, cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = response.json()
            logging.debug(
                "Successfully get projects result: {}".format(result))
        else:
            logging.error(
                "Fail to get projects result, response code: {}, error: {}".
                    format(response.status_code, response.content))
        return result

    # HEAD /projects
    def check_project_exist(self, project_name):
        result = False
        path = '%s://%s/api/projects?project_name=%s' % (self.protocol,
                                                         self.host,
                                                         project_name)
        response = requests.head(
            path, cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = True
            logging.debug(
                "Successfully check project exist, result: {}".format(result))
        elif response.status_code == 404:
            result = False
            logging.debug(
                "Successfully check project exist, result: {}".format(result))
        else:
            logging.error(
                "Fail to check project exist, response code: {}, error: {}".
                    format(response.status_code, response.content))
        return result

    # POST /projects
    def create_project(self, project_name, public=False,
                       enable_content_trust=False, prevent_vulnerable_images_from_running=False,
                       prevent_vulnerable_images_from_running_severity=None,
                       automatically_scan_images_on_push=False):
        result = False
        path = '%s://%s/api/projects' % (self.protocol, self.host)
        request_body = json.dumps({
            'project_name': project_name,
            'public': public,
            'enable_content_trust': enable_content_trust,
            'prevent_vulnerable_images_from_running': prevent_vulnerable_images_from_running,
            'prevent_vulnerable_images_from_running_severity': prevent_vulnerable_images_from_running_severity,
            'automatically_scan_images_on_push': automatically_scan_images_on_push
        })
        response = requests.post(
            path,
            cookies={'beegosessionID': self.session_id},
            data=request_body,
            verify=False)
        if response.status_code == 201:
            result = True
            logging.debug("Successfully create project with project name: {}".
                          format(project_name))
        else:
            logging.error(
                "Fail to create project with project name: {}, response code: {}, error: {}".
                    format(project_name, response.status_code, response.content))
        return result

    # DELETE /projects/{project_id}

    def delete_project(self, project_id):
        result = False
        path = '%s://%s/api/projects/%s' % (self.protocol, self.host,
                                            project_id)
        response = requests.delete(
            path, cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = True
            logging.debug("Successfully delete project with project id: {}".
                          format(project_id))
        else:
            logging.error(
                "Fail to delete project with project id: {}, response code: {}, error: {}".
                    format(project_id, response.status_code, response.content))
        return result

    # GET /projects/{project_id}

    def get_project_info(self, project_id):
        result = None
        path = '%s://%s/api/projects/%s' % (self.protocol, self.host,
                                            project_id)
        response = requests.get(
            path, cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = response.json()
            logging.debug("Successfully get project info: {}".format(result))
        else:
            logging.error(
                "Fail to get project info, response code: {}, error: {}".
                    format(response.status_code, response.content))
        return result

    # PUT /projects/{project_id}/publicity
    def set_project_publicity(self, project_id, is_public):
        result = False
        path = '%s://%s/api/projects/%s/publicity' % (self.protocol, self.host,
                                                      project_id)
        request_body = json.dumps({'public': is_public})
        response = requests.put(
            path,
            cookies={'beegosessionID': self.session_id},
            data=request_body,
            verify=False)
        if response.status_code == 200:
            result = True
            logging.debug("Success to set project id: {} with publicity: {}".
                          format(project_id, is_public))
        else:
            logging.error(
                "Fail to set publicity to project id: {} with status code: {}, error: {}".
                    format(project_id, response.status_code, response.content))
        return result

    # GET /projects/{project_id}/logs
    def get_project_logs(self,
                         project_id,
                         username=None,
                         repo=None,
                         tag=None,
                         operation=None,
                         begin_time=None,
                         end_time=None,
                         page=None,
                         page_size=None):
        result = None
        path = '%s://%s/api/projects/%s/logs' % (self.protocol,
                                                 self.host, project_id)
        payload = {
            'username': username,
            'repository': repo,
            'tag': tag,
            'operation': operation,
            'begin_timestamp': begin_time,
            'end_timestamp': end_time,
            'page': page,
            'page_size': page_size
        }
        response = requests.get(
            path, params=payload,
            cookies={'beegosessionID': self.session_id},
            verify=False)
        if response.status_code == 200:
            result = response.json()
            logging.debug("Successfully get project log: {}".format(result))
        else:
            logging.error(
                "Fail to get project log, response code: {}, error: {}".format(
                    response.status_code, response.content))
        return result

    # GET /projects/{project_id}/members
    def get_project_members(self, project_id):
        result = None
        path = '%s://%s/api/projects/%s/members' % (self.protocol, self.host,
                                                    project_id)
        response = requests.get(
            path, cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = response.json()
            logging.debug(
                "Successfully get project members: {}".format(result))
        else:
            logging.error(
                "Fail to get project members, response code: {}, error: {}".
                    format(response.status_code, response.content))
        return result

    # POST /projects/{project_id}/memebers
    def add_user_project_role(self, project_id, role, user_name):
        result = False
        path = '%s://%s/api/projects/%s/members' % (self.protocol, self.host,
                                                    project_id)
        request_body = json.dumps({'roles': role, 'username': user_name})
        response = requests.post(
            path,
            cookies={'beegosessionID': self.session_id},
            data=request_body,
            verify=False)
        if response.status_code == 200:
            result = True
            logging.debug("Successfully add user project role: {}")
        else:
            logging.error(
                "Fail to add user project role, response code: {}, error: {}".
                    format(response.status_code, response.content))
        return result

    # GET /projects/{project_id}/members/{user_id}
    def get_user_project_role(self, project_id, user_id):
        result = None
        path = '%s://%s/api/projects/%s/members/%s' % (self.protocol,
                                                       self.host, project_id,
                                                       user_id)
        response = requests.get(
            path, cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = response.json()
            logging.debug(
                "Successfully get user project role: {}".format(result))
        else:
            logging.error(
                "Fail to get user project role, response code: {}, error: {}".
                    format(response.status_code, response.content))
        return result

    # PUT /projects/{project_id}/members/{user_id}
    def update_user_project_role(self, project_id, user_id, role, user_name):
        result = False
        path = '%s://%s/api/projects/%s/members/%s' % (self.protocol,
                                                       self.host, project_id,
                                                       user_id)
        request_body = json.dumps({'roles': role, 'username': user_name})
        response = requests.put(
            path,
            cookies={'beegosessionID': self.session_id},
            data=request_body,
            verify=False)
        if response.status_code == 200:
            result = True
            logging.debug("Successfully update user project role.")
        else:
            logging.error(
                "Fail to update user project role with status code: {}, error: {}".
                    format(response.status_code, response.content))
        return result

    # DELETE /projects/{project_id}/members/{user_id}
    def delete_user_project_role(self, project_id, user_id):
        result = False
        path = '%s://%s/api/projects/%s/members/%s' % (self.protocol,
                                                       self.host, project_id,
                                                       user_id)
        response = requests.delete(
            path, cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = True
            logging.debug("Successfully delete user project role")
        else:
            logging.error(
                "Fail to delete user project role, response code: {}, error: {}".
                    format(response.status_code, response.content))
        return result

    # GET /statistics
    def get_statistics(self):
        result = None
        path = '%s://%s/api/statistics' % (self.protocol, self.host)
        response = requests.get(
            path, cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = response.json()
            logging.debug("Successfully get statistics: {}".format(result))
        else:
            logging.error(
                "Fail to get statistics result, response code: {}, error: {}".
                    format(response.status_code, response.content))
        return result

    # GET /users
    def get_users(self, user_name=None, email=None, page=None, page_size=None):
        result = None
        path = '%s://%s/api/users' % (self.protocol, self.host)
        payload = {'username': user_name, 'email': email, 'page': page, 'page_size': page_size}
        response = requests.get(
            path, params=payload, cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = response.json()
            logging.debug("Successfully get users result: {}".format(result))
        else:
            logging.error(
                "Fail to get users result, response code: {}, error: {}".
                    format(response.status_code, response.content))
        return result

    # POST /users
    def create_user(self, username, email, password, realname, comment):
        result = False
        path = '%s://%s/api/users' % (self.protocol, self.host)
        request_body = json.dumps({
            'username': username,
            'email': email,
            'password': password,
            'realname': realname,
            'comment': comment
        })
        response = requests.post(
            path,
            cookies={'beegosessionID': self.session_id},
            data=request_body,
            verify=False)
        if response.status_code == 201:
            result = True
            logging.debug(
                "Successfully create user with username: {}".format(username))
        else:
            logging.error(
                "Fail to create user with username: {}, response code: {}, error: {}".
                    format(username, response.status_code, response.content))
        return result

    # GET /users/current
    def get_current_user(self):
        result = None
        path = '%s://%s/api/users/current' % (self.protocol, self.host)
        response = requests.get(
            path, cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = response.json()
            logging.debug("Successfully get current user: {}".format(result))
        else:
            logging.error(
                "Fail to get current user, response code: {}, error: {}".
                    format(response.status_code, response.content))

        return result

    # GET /users/{user_id}
    def get_user(self, user_id):
        result = None
        path = '%s://%s/api/users/%s' % (self.protocol, self.host, user_id)
        response = requests.get(
            path, cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = response.json()
            logging.debug("Successfully get current user: {}".format(result))
        else:
            logging.error(
                "Fail to get current user, response code: {}, error: {}".
                    format(response.status_code, response.content))
        return result

    # PUT /users/{user_id}
    def update_user_profile(self, user_id, email=None, realname=None, comment=None):
        result = False
        path = '%s://%s/api/users/%s' % (self.protocol, self.host, user_id)
        response = requests.get(path, cookies={'beegosessionID': self.session_id}, verify=False)
        if email == None:
            email = response.json()['email']
        if realname == None:
            realname = response.json()['realname']
        if comment == None:
            comment = response.json()['comment']
        print comment
        request_body = json.dumps({
            'email': email,
            'realname': realname,
            'comment': comment
        })
        response = requests.put(
            path,
            cookies={'beegosessionID': self.session_id},
            data=request_body,
            verify=False)
        if response.status_code == 200:
            result = True
            logging.debug("Successfully update user profile with user id: {}".
                          format(user_id))
        else:
            logging.error(
                "Fail to update user profile with user id: {}, response code: {}, error: {}".
                    format(user_id, response.status_code, response.content))
        return result

    # DELETE /users/{user_id}
    def delete_user(self, user_id):
        result = False
        path = '%s://%s/api/users/%s' % (self.protocol, self.host, user_id)
        response = requests.delete(
            path, cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = True
            logging.debug(
                "Successfully delete user with id: {}".format(user_id))
        else:
            logging.error("Fail to delete user, response code: {}, error: {}".
                          format(response.status_code, response.content))
        return result

    # PUT /users/{user_id}/password
    def change_password(self, user_id, old_password, new_password):
        result = False
        path = '%s://%s/api/users/%s/password' % (self.protocol, self.host,
                                                  user_id)
        request_body = json.dumps({
            'old_password': old_password,
            'new_password': new_password
        })
        response = requests.put(
            path,
            cookies={'beegosessionID': self.session_id},
            data=request_body,
            verify=False)
        if response.status_code == 200:
            result = True
            logging.debug(
                "Successfully change password for user id: {}".format(user_id))
        else:
            logging.error(
                "Fail to change password for user, response code: {}, error: {}".
                    format(response.status_code, response.content))
        return result

    # PUT /users/{user_id}/sysadmin
    def toggle_admin(self, user_id, has_admin_role):
        result = False
        path = '%s://%s/api/users/%s/sysadmin' % (self.protocol, self.host,
                                                  user_id)
        request_body = json.dumps({'has_admin_role': has_admin_role})
        response = requests.put(
            path,
            cookies={'beegosessionID': self.session_id},
            data=request_body,
            verify=False)
        if response.status_code == 200:
            result = True
            logging.debug("Successfully toggle user admin with user id: {}".
                          format(user_id))
        else:
            logging.error(
                "Fail to toggle user admin with user id: {}, response code: {}, error: {}".
                    format(user_id, response.status_code, response.content))
        return result

    # GET /repositories
    def get_repositories(self,
                         project_id,
                         repo_name=None,
                         page=None,
                         page_size=None):
        result = None
        path = '%s://%s/api/repositories' % (self.protocol,
                                             self.host)
        payload = {'project_id': project_id,
                   'q': repo_name,
                   'page': page,
                   'page_size': page_size
                   }
        response = requests.get(
            path, params=payload, cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = response.json()
            logging.debug(
                "Successfully get repositories with id: {}, result: {}".format(
                    project_id, result))
        else:
            logging.error(
                "Fail to get repositories, response code: {}, error: {}".
                    format(response.status_code, response.content))
        return result

    # DELETE /repositories{repo_name}
    def delete_repository(self, repo_name):
        result = False
        path = '%s://%s/api/repositories/%s' % (self.protocol, self.host,
                                                repo_name)
        response = requests.delete(
            path, cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = True
            logging.debug("Successfully delete the repository {}".
                          format(repo_name))
        else:
            logging.error(
                "Fail to delete the repository {}, response code: {}, error: {}".
                    format(repo_name, response.status_code, response.content))
        return result

    # GET /repositories{repo_name}/tags/{tag}
    def get_repository_tag(self, repo_name, tag):
        result = None
        path = '%s://%s/api/repositories/%s/tags/%s' % (self.protocol,
                                                        self.host, repo_name,
                                                        tag)
        response = requests.get(
            path, cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = response.json()
            logging.debug("Successfully get {} tag {}".
                          format(repo_name, tag))
        else:
            logging.error(
                "Fail to get {} tag {}, response code: {}, error: {}". \
                    format(repo_name, tag, response.status_code, response.content))
        return result

    # DELETE /repositories{repo_name}/tags/{tag}
    def delete_repository_tag(self, repo_name, tag):
        result = False
        path = '%s://%s/api/repositories/%s/tags/%s' % (self.protocol,
                                                        self.host, repo_name,
                                                        tag)
        response = requests.delete(
            path, cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = True
            logging.debug("Successfully delete repository {} with tag {}".
                          format(repo_name, tag))
        else:
            logging.error(
                "Fail to delete repository {} with tag {}, response code: {}, error: {}".
                    format(repo_name, tag, response.status_code, response.content))
        return result

    # Get /repositories/{repo_name}/tags
    def get_repository_tags(self, repo_name):
        result = None
        path = '%s://%s/api/repositories/%s/tags' % (self.protocol, self.host,
                                                     repo_name)
        response = requests.get(
            path, cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = response.json()
            logging.debug(
                "Successfully get tags with repo name: {}, result: {}".format(
                    repo_name, result))
        else:
            logging.error(
                "Fail to get tags with repo name: {}, response code: {}, error: {}".
                    format(repo_name, response.status_code, response.content))
        return result

    # GET /repositories/{repo_name}/tags/{tag}/manifest
    def get_repository_manifest(self, repo_name, tag, version=None):
        result = None
        path = '%s://%s/api/repositories/%s/tags/%s/manifest' % (self.protocol,
                                                                 self.host,
                                                                 repo_name,
                                                                 tag)
        if version:
            path += '?version=%s' % (version)
        response = requests.get(
            path, cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = response.json()
            logging.debug(
                "Successfully get manifest with repo name: {}, tag: {}, result: {}".
                    format(repo_name, tag, result))
        else:
            logging.error(
                "Fail to get manifest with repo name: {}, tag: {}, response code: {}, error: {}".
                    format(repo_name, tag, response.status_code, response.content))
        return result

    # POST /repositories/{repo_name}/tags/{tag}/scan
    def create_scan_job(self, repo_name, tag):
        result = False
        path = '%s://%s/api/repositories/%s/tags/%s/scan' % (self.protocol, self.host, repo_name, tag)
        response = requests.post(path, cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = True
            logging.debug("Successfully create scan job with repo name: {}, tag: {}".format(repo_name, tag))
        else:
            logging.error(
                "Fail to create scan job with repo name: {}, tag: {}, response code: {}, error: {}".format(repo_name,
                                                                                                           tag,
                                                                                                           response.status_code,
                                                                                                           response.content))
        return result

    # GET /repositories/{repo_name}/tags/{tag}/vulnerability/details
    def get_scan_detail(self, repo_name, tag):
        result = None
        path = '%s://%s/api/repositories/%s/tags/%s/vulnerability/details' % (self.protocol, self.host, repo_name, tag)
        response = requests.get(path, cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = response.json()
            logging.debug("Successfully get scan detail with repo name: {}, tag: {}".format(repo_name, tag))
        else:
            logging.error(
                "Fail to get scan detail with repo name: {}, tag: {}, response code: {}, error: {}".format(
                    repo_name,
                    tag,
                    response.status_code,
                    response.content))
        return result

    # GET /repositories/{repo_name}/signatures
    def get_repository_signatures(self, repo_name):
        result = None
        path = '%s://%s/api/repositories/%s/signatures' % (self.protocol,
                                                           self.host,
                                                           repo_name)
        response = requests.get(
            path, cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = response.json()
            logging.debug(
                "Successfully get signatures with repo name: {}, result: {}".
                    format(repo_name, result))
        else:
            logging.error(
                "Fail to get signatures with repo name: {}, response code: {}, error: {}".
                    format(repo_name, response.status_code, response.content))
        return result

    # GET /repositories/top
    def get_top_repositories(self, count=None):
        result = None
        path = '%s://%s/api/repositories/top' % (self.protocol, self.host)
        if count:
            path += "?count=%s" % (count)
        response = requests.get(
            path, cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = response.json()
            logging.debug(
                "Successfully get top accessed repositories, result: {}".
                    format(result))
        else:
            logging.error(
                "Fail to get top accessed repositories, response code: {}, error: {}".
                    format(response.status_code, response.content))
        return result

    # GET /logs
    def get_logs(self, username=None, repository=None, tag=None, operation=None, begin_timestamp=None,
                 end_timestamp=None, page=None, page_size=None):
        result = None
        path = '%s://%s/api/logs' % (self.protocol, self.host)
        payload = {'username': username,
                   'repository': repository,
                   'tag': tag,
                   'operation': operation,
                   'begin_timestamp': begin_timestamp,
                   'end_timestamp': end_timestamp,
                   'page': page,
                   'page_size': page_size}
        response = requests.get(
            path, params=payload, cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = response.json()
            logging.debug("Successfully get logs")
        else:
            logging.error("Fail to get logs and response code: {}, error: {}".
                          format(response.status_code, response.content))
        return result

    # GET /jobs/replication
    def get_jobs(self,
                 policy_id,
                 num=None,
                 start_time=None,
                 end_time=None,
                 repository=None,
                 status=None,
                 page=None,
                 page_size=None):
        result = None
        path = '%s://%s/api/jobs/replication' % (self.protocol, self.host)
        payload = {'policy_id': policy_id,
                   'num': num,
                   'start_time': start_time,
                   'end_time': end_time,
                   'repository': repository,
                   'status': status,
                   'page': page,
                   'page_size': page_size}
        response = requests.get(
            path, params=payload, cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = response.json()
            logging.debug("Successfully get jobs")
        else:
            logging.error("Fail to get jobs, response code: {}, error: {}".
                          format(response.status_code, response.content))

        return result

    # DELETE /jobs/replication/{id}
    def delete_job(self, job_id):
        result = False
        path = '%s://%s/api/jobs/replication/%s' % (self.protocol, self.host,
                                                    job_id)
        response = requests.delete(
            path, cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = True
            logging.debug("Successfully delete job with id: {}".format(job_id))
        else:
            logging.error(
                "Fail to delete job with id: {} and response code: {}, error: {}".
                    format(job_id, response.status_code, response.content))
        return result

    # GET /jobs/replication/{id}/log
    def get_job_logs(self, job_id):
        result = None
        path = '%s://%s/api/jobs/replication/%s/log' % (self.protocol,
                                                        self.host, job_id)
        response = requests.get(
            path, cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = response.content
            logging.debug("Successfully get job logs")
        else:
            logging.error(
                "Fail to get job logs and response code: {}, error: {}".format(
                    response.status_code, response.content))
        return result

    # GET /policies/replication
    def get_policies(self, name=None, project_id=None):
        result = None
        path = '%s://%s/api/policies/replication' % (self.protocol, self.host)
        payload = {'name': name,
                   'project_id': project_id}
        response = requests.get(
            path, params=payload, cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = response.json()
            logging.debug(
                "Successfully get policies, result: {}".format(result))
        else:
            logging.error(
                "Fail to get policies with response code: {}, error: {}".
                    format(response.status_code, response.content))
        return result

    # POST /policies/replication
    def create_policy(self, project_id, target_id, name, enable):
        result = False
        path = '%s://%s/api/policies/replication' % (self.protocol, self.host)
        request_body = json.dumps({
            'project_id': project_id,
            'target_id': target_id,
            'name': name,
            'enabled': enable
        })
        response = requests.post(
            path,
            cookies={'beegosessionID': self.session_id},
            data=request_body,
            verify=False)
        if response.status_code == 201:
            result = True
            logging.debug(
                "Successfully create policy with name: {}".format(name))
        else:
            logging.error(
                "Fail to create policy with name: {}, response code: {}, error: {}".
                    format(name, response.status_code, response.content))
        return result

    # GET /policies/replication/{id}
    def get_policy(self, policy_id):
        result = None
        path = '%s://%s/api/policies/replication/%s' % (self.protocol,
                                                        self.host, policy_id)
        response = requests.get(
            path, cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = response.json()
            logging.debug("Successfully get policy info")
        else:
            logging.error(
                "Fail to get policy info and response code: {}, error: {}".
                    format(response.status_code, response.content))
        return result

    # PUT /policies/replication/{id}
    def update_policy(self, policy_id, target_id=None, name=None, enabled=None, description=None,
                      cron_str=None):
        result = False
        path = '%s://%s/api/policies/replication/%s' % (self.protocol,
                                                        self.host, policy_id)
        response = requests.get(
            path, cookies={'beegosessionID': self.session_id}, verify=False)
        if target_id == None:
            target_id = response.json()['target_id']
        if name == None:
            name = response.json()['name']
        if enabled == None:
            enabled = response.json()['enabled']
        if description == None:
            description = response.json()['description']
        if cron_str == None:
            cron_str = response.json()['cron_str']
        request_body = json.dumps({
            'target_id': target_id,
            'name': name,
            'enabled': enabled,
            'description': description,
            'cron_str': cron_str
        })
        response = requests.put(
            path,
            cookies={'beegosessionID': self.session_id},
            data=request_body,
            verify=False)
        if response.status_code == 200:
            result = True
            logging.debug("Successfully update policy for policy id: {}".
                          format(policy_id))
        else:
            logging.error(
                "Fail to update policy for policy id: {} with response code: {}, error: {}".
                    format(policy_id, response.status_code, response.content))
        return result

    # PUT /policies/replication/{id}/enablement
    def update_policy_enablement(self, policy_id, enabled):
        result = False
        path = '%s://%s/api/policies/replication/%s/enablement' % (
            self.protocol, self.host, policy_id)
        request_body = json.dumps({'enabled': enabled})
        response = requests.put(
            path,
            cookies={'beegosessionID': self.session_id},
            data=request_body,
            verify=False)
        if response.status_code == 200:
            result = True
            logging.debug(
                "Successfully update policy enablement for policy id: {}".
                    format(policy_id))
        else:
            logging.error(
                "Fail to update policy enablement for policy id: {} with response code: {}, error: {}".
                    format(policy_id, response.status_code, response.content))
        return result

    # GET /targets
    def get_replication_targets(self, name=None):
        result = None
        path = '%s://%s/api/targets' % (self.protocol, self.host)
        if name:
            path += "?name=%s" % (name)
        response = requests.get(
            path, cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = response.json()
            logging.debug("Successfully get targets")
        else:
            logging.error(
                "Fail to get targets and response code: {}, error: {}".format(
                    response.status_code, response.content))
        return result

    # POST /targets
    def create_replication_target(self, endpoint, name, user_name, password):
        result = False
        path = '%s://%s/api/targets' % (self.protocol, self.host)
        request_body = json.dumps({
            'endpoint': endpoint,
            'name': name,
            'username': user_name,
            'password': password
        })
        response = requests.post(
            path,
            cookies={'beegosessionID': self.session_id},
            data=request_body,
            verify=False)
        if response.status_code == 201:
            result = True
            logging.debug(
                "Successfully create replication target with name: {}".format(
                    name))
        else:
            logging.error(
                "Fail to create replication target with name: {}, response code: {}, error: {}".
                    format(name, response.status_code, response.content))
        return result

    # POST /targets/ping
    def ping_replication_target(self, endpoint, user_name, password):
        result = False
        path = '%s://%s/api/targets/ping' % (self.protocol, self.host)
        request_body = json.dumps({
            'endpoint': endpoint,
            'username': user_name,
            'password': password
        })
        response = requests.post(
            path,
            cookies={'beegosessionID': self.session_id},
            data=request_body,
            verify=False)
        if response.status_code == 200:
            result = True
            logging.debug("Successfully ping replication target ")
        else:
            logging.error(
                "Fail to ping replication target, response code: {}, error: {}".
                    format(response.status_code, response.content))
        return result

    # POST /targets/{id}/ping
    def ping_replication_target_with_id(self, target_id):
        result = False
        path = '%s://%s/api/targets/%s/ping' % (self.protocol, self.host,
                                                target_id)
        response = requests.post(
            path, cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = True
            logging.debug("Successfully ping replication target with id: {} ".
                          format(target_id))
        else:
            logging.error(
                "Fail to ping replication target with id: {}, response code: {}, error: {}".
                    format(target_id, response.status_code, response.content))
        return result

    # POST /targets/{id}
    def update_replication_target(self, target_id, endpoint=None, name=None, user_name=None,
                                  password=None):
        result = False
        path = '%s://%s/api/targets/%s' % (self.protocol, self.host, target_id)
        response = requests.get(
            path, cookies={'beegosessionID': self.session_id}, verify=False)
        if endpoint == None:
            endpoint = response.json()['endpoint']
        if name == None:
            name = response.json()['name']
        if user_name == None:
            user_name = response.json()['username']
        if password == None:
            password = response.json()['password']
        request_body = json.dumps({
            'endpoint': endpoint,
            'name': name,
            'username': user_name,
            'password': password
        })
        response = requests.put(
            path,
            cookies={'beegosessionID': self.session_id},
            data=request_body,
            verify=False)
        if response.status_code == 200:
            result = True
            logging.debug(
                "Successfully update replication target with name: {}".format(
                    name))
        else:
            logging.error(
                "Fail to update replication target with name: {}, response code: {}, error: {}".
                    format(name, response.status_code, response.content))
        return result

    # GET /targets/{id}
    def get_replication_target(self, target_id):
        result = None
        path = '%s://%s/api/targets/%s' % (self.protocol, self.host, target_id)
        response = requests.get(
            path, cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = response.json()
            logging.debug("Successfully get target")
        else:
            logging.error(
                "Fail to get target and response code: {}, error: {}".format(
                    response.status_code, response.content))
        return result

    # DELETE /targets/{id}
    def delete_replication_target(self, target_id):
        result = None
        path = '%s://%s/api/targets/%s' % (self.protocol, self.host, target_id)
        response = requests.delete(
            path, cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = True
            logging.debug("Successfully delete replication target")
        else:
            logging.error(
                "Fail to delete replication target and response code: {}, error: {}".
                    format(response.status_code, response.content))
        return result

    # GET /targets/{id}/policies
    def get_replication_target_policies(self, target_id):
        result = None
        path = '%s://%s/api/targets/%s/policies' % (self.protocol, self.host,
                                                    target_id)
        response = requests.get(
            path, cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = response.json()
            logging.debug("Successfully get target policies")
        else:
            logging.error(
                "Fail to get target policies and response code: {}, error: {}".
                    format(response.status_code, response.content))
        return result

    # POST /internal/syncregistry
    def sync_registry(self):
        result = False
        path = '%s://%s/api/internal/syncregistry' % (self.protocol, self.host)
        response = requests.post(
            path, cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = True
            logging.debug("Successfully sync registry to DB")
        else:
            logging.error(
                "Fail to sync registry to DB, response code: {}, error: {}".
                    format(response.status_code, response.content))

        return result

    # GET /systeminfo
    def get_systeminfo(self):
        result = None
        path = '%s://%s/api/systeminfo' % (self.protocol, self.host)
        response = requests.get(
            path, cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = response.json()
            logging.debug("Successfully get system info")
        else:
            logging.error(
                "Fail to get system info, response code: {}, error: {}".format(
                    response.status_code, response.content))
        return result

    # GET /systeminfo/volumes
    def get_systeminfo_volumes(self):
        result = None
        path = '%s://%s/api/systeminfo/volumes' % (self.protocol, self.host)
        response = requests.get(
            path, cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = response.json()
            logging.debug("Successfully get system volumes info")
        else:
            logging.error(
                "Fail to get system volumes info, response code: {}, error: {}".
                    format(response.status_code, response.content))
        return result

    # GET /systeminfo/getcert
    def get_systeminfo_cert(self):
        result = False
        path = '%s://%s/api/systeminfo/getcert' % (self.protocol, self.host)
        response = requests.get(
            path, cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = True
            logging.debug("Successfully get system cert")
        else:
            logging.error(
                "Fail to get system cert, response code: {}, error: {}".format(
                    response.status_code, response.content))
        return result

    # POST /ldap/ping
    def ping_ldap(self,
                  ldap_url=None,
                  ldap_search_dn=None,
                  ldap_search_password=None,
                  ldap_base_dn=None,
                  ldap_filter=None,
                  ldap_uid=None,
                  ldap_scope=None,
                  ldap_connection_timeout=None):
        result = False
        response = None
        path = '%s://%s/api/ldap/ping' % (self.protocol, self.host)
        if ldap_url:
            request_body = json.dumps({
                'ldap_url':
                    ldap_url,
                'ldap_search_dn':
                    ldap_search_dn,
                'ldap_search_password':
                    ldap_search_password,
                'ldap_base_dn':
                    ldap_base_dn,
                'ldap_filter':
                    ldap_filter,
                'ldap_uid':
                    ldap_uid,
                'ldap_scope':
                    ldap_scope,
                'ldap_connection_timeout':
                    ldap_connection_timeout
            })
            response = requests.post(
                path,
                cookies={'beegosessionID': self.session_id},
                data=request_body,
                verify=False)
        else:
            response = requests.post(
                path,
                cookies={'beegosessionID': self.session_id},
                verify=False)
        if response.status_code == 200:
            result = True
            logging.debug("Successfully ping ldap service ")
        else:
            logging.error(
                "Fail to ping ldap service, response code: {}, error: {}".
                    format(response.status_code, response.content))
        return result

    # POST /ldap/users/search
    def search_ldap_users(self,
                          user_name=None,
                          ldap_url=None,
                          ldap_search_dn=None,
                          ldap_search_password=None,
                          ldap_base_dn=None,
                          ldap_filter=None,
                          ldap_uid=None,
                          ldap_scope=None,
                          ldap_connection_timeout=None):
        result = None
        response = None
        path = '%s://%s/api/ldap/users/search' % (self.protocol, self.host)
        if user_name:
            path += '?username=%s' % (user_name)
        if ldap_url:
            request_body = json.dumps({
                'ldap_url':
                    ldap_url,
                'ldap_search_dn':
                    ldap_search_dn,
                'ldap_search_password':
                    ldap_search_password,
                'ldap_base_dn':
                    ldap_base_dn,
                'ldap_filter':
                    ldap_filter,
                'ldap_uid':
                    ldap_uid,
                'ldap_scope':
                    ldap_scope,
                'ldap_connection_timeout':
                    ldap_connection_timeout
            })
            response = requests.post(
                path,
                cookies={'beegosessionID': self.session_id},
                data=request_body,
                verify=False)
        else:
            response = requests.post(
                path,
                cookies={'beegosessionID': self.session_id},
                verify=False)
        if response.status_code == 200:
            result = response.json()
            logging.debug("Successfully search ldap users ")
        else:
            logging.error(
                "Fail to search ldap users, response code: {}, error: {}".
                    format(response.status_code, response.content))
        return result

    # POST /ldap/users/import
    def import_ldap_users(self, ldap_uid_list):
        result = False
        path = '%s://%s/api/ldap/users/import' % (self.protocol, self.host)

        request_body = json.dumps({'ldap_uid_list': ldap_uid_list})
        response = requests.post(
            path,
            cookies={'beegosessionID': self.session_id},
            data=request_body,
            verify=False)
        if response.status_code == 200:
            result = True
            logging.debug("Successfully import ldap users ")
        else:
            # TODO:when error is 500, content[36:] are messy code.
            if response.status_code == 500:
                logging.error(
                    "Fail to import ldap users, response code: {}, error: {}".
                        format(response.status_code, response.content[0:36]))
            else:
                logging.error(
                    "Fail to import ldap users, response code: {}, error: {}".
                        format(response.status_code, response.content))
        return result

    # GET /configurations
    def get_configurations(self):
        result = None
        path = '%s://%s/api/configurations' % (self.protocol, self.host)
        response = requests.get(
            path, cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = response.json()
            logging.debug("Successfully get system configurations")
        else:
            logging.error(
                "Fail to get system configurations, response code: {}, error: {}".
                    format(response.status_code, response.content))
        return result

    # PUT /configurations
    def update_configurations(self, auth_mode=None, email_from=None, email_host=None,
                              email_identity=None, email_password=None, email_port=None,
                              email_ssl=None, email_username=None, ldap_url=None,
                              ldap_base_dn=None, ldap_search_dn=None, ldap_search_password=None,
                              ldap_filter=None, ldap_uid=None, ldap_scope=None, ldap_timeout=None,
                              project_creation_restriction=None,
                              self_registration=None, verify_remote_cert=None):
        result = False
        path = '%s://%s/api/configurations' % (self.protocol, self.host)
        payload = {}
        if auth_mode != None:
            payload['auth_mode'] = auth_mode
        if email_from != None:
            payload['email_from'] = email_from
        if email_host != None:
            payload['email_host'] = email_host
        if email_identity != None:
            payload['email_identity'] = email_identity
        if email_password != None:
            payload['email_password'] = email_password
        if email_port != None:
            payload['email_port'] = email_port
        if email_ssl != None:
            payload['email_ssl'] = email_ssl
        if email_username != None:
            payload['email_username'] = email_username
        if ldap_url != None:
            payload['ldap_url'] = ldap_url
        if ldap_base_dn != None:
            payload['ldap_base_dn'] = ldap_base_dn
        if ldap_search_dn != None:
            payload['ldap_search_dn'] = ldap_search_dn
        if ldap_search_password != None:
            payload['ldap_search_password'] = ldap_search_password
        if ldap_filter != None:
            payload['ldap_filter'] = ldap_filter
        if ldap_scope != None:
            payload['ldap_scope'] = ldap_scope
        if ldap_uid != None:
            payload['ldap_uid'] = ldap_uid
        if ldap_timeout != None:
            payload['ldap_timeout'] = ldap_timeout
        if project_creation_restriction != None:
            payload['project_creation_restriction'] = project_creation_restriction
        if self_registration != None:
            payload['self_registration'] = self_registration
        if verify_remote_cert != None:
            payload['verify_remote_cert'] = verify_remote_cert

        request_body = json.dumps(payload)
        response = requests.put(
            path,
            cookies={'beegosessionID': self.session_id},
            data=request_body,
            verify=False)
        if response.status_code == 200:
            result = True
            logging.debug("Successfully update system configurations")
        else:
            logging.error(
                "Fail to update system configurations with response code {}, error: {}".
                    format(response.status_code, response.content))
        return result

    # POST /configurations/reset
    def reset_configurations(self):
        result = False
        path = '%s://%s/api/configurations/reset' % (self.protocol, self.host)
        response = requests.post(
            path, cookies={'beegosessionID': self.session_id}, verify=False)
        if response.status_code == 200:
            result = True
            logging.debug("Successfully reset system configurations ")
        else:
            logging.error(
                "Fail to reset system configurations, response code: {}, error: {}".
                    format(response.status_code, response.content))
        return result

    # POST /email/ping
    def ping_email(self, email_host=None, email_port=None, email_username=None, email_password=None,
                   email_ssl=None, email_identity=None, ):
        result = False
        response = None
        path = '%s://%s/api/email/ping' % (self.protocol, self.host)
        if email_host:
            request_body = json.dumps({
                'email_host': email_host,
                'email_identity': email_identity,
                'email_password': email_password,
                'email_port': email_port,
                'email_ssl': email_ssl,
                'email_username': email_username
            })
            response = requests.post(
                path,
                cookies={'beegosessionID': self.session_id},
                data=request_body,
                verify=False)
        else:
            response = requests.post(
                path,
                cookies={'beegosessionID': self.session_id},
                verify=False)
        if response.status_code == 200:
            result = True
            logging.debug("Successfully ping email server ")
        else:
            logging.error(
                "Fail to ping email server, response code: {}, error: {}".
                    format(response.status_code, response.content))
        return result
