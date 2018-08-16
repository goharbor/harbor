# -*- coding: utf-8 -*-

import sys
import base
import swagger_client

class Repository(base.Base):
    def list_tags(self, repository, **kwargs):
        client = self._get_client(**kwargs)
        return client.repositories_repo_name_tags_get(repository)
    
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