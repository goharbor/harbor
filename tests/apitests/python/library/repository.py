# -*- coding: utf-8 -*-

import time
import base
import swagger_client
from docker_api import DockerAPI
from swagger_client.rest import ApiException

def pull_harbor_image(registry, username, password, image, tag, expected_error_message = None):
    _docker_api = DockerAPI()
    _docker_api.docker_login(registry, username, password)
    time.sleep(2)
    _docker_api.docker_image_pull(r'{}/{}'.format(registry, image), tag = tag, expected_error_message = expected_error_message)

def push_image_to_project(project_name, registry, username, password, image, tag, expected_error_message = None):
    _docker_api = DockerAPI()
    _docker_api.docker_login(registry, username, password)
    time.sleep(2)

    _docker_api.docker_image_pull(image, tag = tag)
    time.sleep(2)

    new_harbor_registry, new_tag = _docker_api.docker_image_tag(r'{}:{}'.format(image, tag), r'{}/{}/{}'.format(registry, project_name, image))
    time.sleep(2)

    _docker_api.docker_image_push(new_harbor_registry, new_tag, expected_error_message = expected_error_message)

    return r'{}/{}'.format(project_name, image), new_tag

def is_repo_exist_in_project(repositories, repo_name):
    result = False
    for reop in repositories:
        if reop.name == repo_name:
            return True
    return result

class Repository(base.Base):

    def list_tags(self, repository, **kwargs):
        client = self._get_client(**kwargs)
        return client.repositories_repo_name_tags_get(repository)

    def get_tag(self, repo_name, tag, **kwargs):
        client = self._get_client(**kwargs)
        return client.repositories_repo_name_tags_tag_get(repo_name, tag)

    def image_exists(self, repository, tag, **kwargs):
        tags = self.list_tags(repository, **kwargs)
        exist = False
        for t in tags:
            if t.name == tag:
                exist = True
                break
        return exist

    def image_should_exist(self, repository, tag, **kwargs):
        if not self.image_exists(repository, tag, **kwargs):
            raise Exception("image %s:%s not exist" % (repository, tag))

    def image_should_not_exist(self, repository, tag, **kwargs):
        if self.image_exists(repository, tag, **kwargs):
            raise Exception("image %s:%s exists" % (repository, tag))

    def delete_repoitory(self, repo_name, **kwargs):
        client = self._get_client(**kwargs)
        _, status_code, _ = client.repositories_repo_name_delete_with_http_info(repo_name)
        base._assert_status_code(200, status_code)

    def get_repository(self, project_id, **kwargs):
        client = self._get_client(**kwargs)
        data, status_code, _ = client.repositories_get_with_http_info(project_id)
        base._assert_status_code(200, status_code)
        return data

    def add_label_to_tag(self, repo_name, tag, label_id, expect_status_code = 200, **kwargs):
        client = self._get_client(**kwargs)
        label = swagger_client.Label(id=label_id)
        _, status_code, _ = client.repositories_repo_name_tags_tag_labels_post_with_http_info(repo_name, tag, label)
        base._assert_status_code(expect_status_code, status_code)

    def get_repo_signatures(self, repo_name, expect_status_code = 200, **kwargs):
        client = self._get_client(**kwargs)
        data, status_code, _ = client.repositories_repo_name_signatures_get_with_http_info(repo_name)
        base._assert_status_code(expect_status_code, status_code)
        return data

    def check_image_not_scanned(self, repo_name, tag, **kwargs):
        tag = self.get_tag(repo_name, tag, **kwargs)
        if tag.scan_overview != None:
            raise Exception("Image should be <Not Scanned> state!")

    def check_image_scan_result(self, repo_name, tag, expected_scan_status = "finished", **kwargs):
        timeout_count = 30
        while True:
            time.sleep(5)
            timeout_count = timeout_count - 1
            if (timeout_count == 0):
                break
            _tag = self.get_tag(repo_name, tag, **kwargs)
            if _tag.name == tag and _tag.scan_overview !=None:
                if _tag.scan_overview.scan_status == expected_scan_status:
                    return
        raise Exception("Scan image result is not as expected {}.".format(expected_scan_status))

    def scan_image(self, repo_name, tag, expect_status_code = 200, **kwargs):
        client = self._get_client(**kwargs)
        data, status_code, _ = client.repositories_repo_name_tags_tag_scan_post_with_http_info(repo_name, tag)
        base._assert_status_code(expect_status_code, status_code)
        return data

    def repository_should_exist(self, project_id, repo_name, **kwargs):
        repositories = self.get_repository(project_id, **kwargs)
        if is_repo_exist_in_project(repositories, repo_name) == False:
            raise Exception("Repository {} is not exist.".format(repo_name))

    def signature_should_exist(self, repo_name, tag, **kwargs):
        signatures = self.get_repo_signatures(repo_name, **kwargs)
        for each_sign in signatures:
            if each_sign.tag == tag and len(each_sign.hashes["sha256"]) == 44:
                print "sha256:", len(each_sign.hashes["sha256"])
                return
        raise Exception(r"Signature of {}:{} is not exist!".format(repo_name, tag))

    def retag_image(self, repo_name, tag, src_image, override=True, expect_status_code = 200, expect_response_body = None, **kwargs):
        client = self._get_client(**kwargs)
        request = swagger_client.RetagReq(tag=tag, src_image=src_image, override=override)

        try:
            data, status_code, _ = client.repositories_repo_name_tags_post_with_http_info(repo_name, request)
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
            if expect_response_body is not None:
                base._assert_status_body(expect_response_body, e.body)
            return

        base._assert_status_code(expect_status_code, status_code)
        base._assert_status_code(200, status_code)
        return data