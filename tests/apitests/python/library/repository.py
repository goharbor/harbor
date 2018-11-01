# -*- coding: utf-8 -*-

import sys
import time
import base
import swagger_client
import docker
try:
    import docker
except ImportError:
    import pip
    pip.main(['install', 'docker'])
    import docker

class Repository(base.Base):
    DCLIENT = docker.APIClient(base_url='unix://var/run/docker.sock',version='auto',timeout=10)

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
        data, status_code, _ = client.repositories_repo_name_tags_tag_labels_post_with_http_info(repo_name, tag, label)
        base._assert_status_code(expect_status_code, status_code)

    def create_repository(self, project_name, registry = None, username = None, password = None, tag = None, image = None):      
        _registry = '127.0.0.1'
        _username = 'admin'
        _password ='Harbor12345'
        _tag = "latest"
        _image ="hello-world"

        if registry is not None:
            _registry = registry
        if username is not None:
            _username = username
        if password is not None:
            _password = password
        if tag is not None:
            _tag = tag
        if image is not None:
            _image = image

        self.docker_login(_registry, _username, _password)
        time.sleep(2)

        self.docker_image_pull(_image, _tag)
        time.sleep(2)

        new_harbor_registry, new_tag = self.docker_image_tag(_image, r'{}/{}/{}'.format(_registry, project_name, _image))
        time.sleep(2)

        self.docker_image_push(new_harbor_registry, new_tag)

        return r'{}/{}'.format(project_name, _image), new_tag

    def docker_login(self, registry, username, password):
        try:
            Repository.DCLIENT.login(registry = registry, username=username, password=password)
        except docker.errors.APIError, e:
            raise Exception(r" Docker login failed, error is [{}]".format (e.message))        

    def docker_image_pull(self, image, tag = None):
        _tag = "latest"
        if tag is not None:
            _tag = tag
        try:
            tag = base._random_name("tag")
            pull_ret = base._get_string_from_unicode(Repository.DCLIENT.pull('{}:{}'.format(image, _tag)))
            print "pull_ret:", pull_ret
        except docker.errors.APIError, e:
            raise Exception(r" Docker pull image {} failed, error is [{}]".format (image, e.message))

    def docker_image_tag(self, image, harbor_registry, tag = None):
        _tag = base._random_name("tag")
        if tag is not None:
            _tag = tag        
        try:
            tag_ret = Repository.DCLIENT.tag(image, harbor_registry, _tag, force=True)
            print "tag_ret:", tag_ret
            return harbor_registry, _tag
        except docker.errors.APIError, e:
            raise Exception(r" Docker tag image {} failed, error is [{}]".format (image, e.message))

    def docker_image_push(self, harbor_registry, tag):
        try:
            push_ret = base._get_string_from_unicode(Repository.DCLIENT.push(harbor_registry, tag, stream=True))
            print "push_ret:", push_ret
        except docker.errors.APIError, e:
            raise Exception(r" Docker tag image {} failed, error is [{}]".format (image, e.message))
        