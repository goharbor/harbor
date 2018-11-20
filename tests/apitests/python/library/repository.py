# -*- coding: utf-8 -*-

import time
import base
import swagger_client
from docker_api import DockerAPI

def push_image_to_project(project_name, registry, username, password, image, tag):
    _docker_api = DockerAPI()
    _docker_api.docker_login(registry, username, password)
    time.sleep(2)

    _docker_api.docker_image_pull(image, tag)
    time.sleep(2)

    new_harbor_registry, new_tag = _docker_api.docker_image_tag(image, r'{}/{}/{}'.format(registry, project_name, image))
    time.sleep(2)

    _docker_api.docker_image_push(new_harbor_registry, new_tag)

    return r'{}/{}'.format(project_name, image), new_tag

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

    def get_not_scanned_image_init_state_success(self, repo_name, tag, **kwargs):
        tag = self.get_tag(repo_name, tag, **kwargs)
        if tag.scan_overview != None:
            raise Exception("Image should be <Not Scanned> state!")

    def check_image_scan_result(self, repo_name, tag, expected_scan_status = "finished", **kwargs):
        scan_finish = False
        timeout_count = 20
        actual_scan_status = "NULL"
        while not (scan_finish):
            time.sleep(5)
            _tag = self.get_tag(repo_name, tag, **kwargs)
            scan_result = False
            print "t.name:{}  tag:{}, t.scan_overview.scan_status:{}".format(_tag.name, tag, _tag.scan_overview.scan_status)
            if _tag.name == tag and _tag.scan_overview !=None:
                if _tag.scan_overview.scan_status == expected_scan_status:
                    scan_finish = True
                    scan_result = True
                    break
                else:
                    actual_scan_status = _tag.scan_overview.scan_status
            timeout_count = timeout_count - 1
            if (timeout_count == 0):
                scan_finish = True
        if not (scan_result):
            raise Exception("Scan image result is not as expected {} actual scan status is {}".format(expected_scan_status, actual_scan_status))

    def scan_image(self, repo_name, tag, expect_status_code = 200, **kwargs):
        client = self._get_client(**kwargs)
        data, status_code, _ = client.repositories_repo_name_tags_tag_scan_post_with_http_info(repo_name, tag)
        base._assert_status_code(expect_status_code, status_code)
        return data

    def scan_not_scanned_image_success(self, repo_name, tag, **kwargs):
        self.get_not_scanned_image_init_state_success(repo_name, tag, **kwargs)
        self.scan_image(repo_name, tag, **kwargs)
        self.check_image_scan_result(repo_name, tag, **kwargs)

