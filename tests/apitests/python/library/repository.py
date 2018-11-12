# -*- coding: utf-8 -*-

import time
import base
import swagger_client
from docker_api import DockerAPI

def create_repository(project_name, registry, username, password, image, tag):
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
        return client.repositories_repo_name_tags_get_with_http_info(repository)

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

    def signature_should_exist(self, repo_name, tag, **kwargs):
        signatures = self.get_repo_signatures(repo_name, **kwargs)
        for each_sign in signatures:
            if each_sign.tag == tag and len(each_sign.hashes["sha256"]) == 44:
                print "sha256:", len(each_sign.hashes["sha256"])
                return
        raise Exception(r"Signature of {}:{} is not exist!".format(repo_name, tag))

