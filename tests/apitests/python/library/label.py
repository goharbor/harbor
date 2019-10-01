# -*- coding: utf-8 -*-

import sys
import base
import swagger_client

class Label(base.Base):
    def create_label(self, name=None, desc="",
        color="", scope="g", project_id=0, **kwargs):
        if name is None:
            name = base._random_name("label")
        label = swagger_client.Label(name=name,
            description=desc, color=color,
            scope=scope, project_id=project_id)
        client = self._get_client(**kwargs)
        _, _, header = client.labels_post_with_http_info(label)
        return base._get_id_from_header(header), name

    def add_label_to_image(self, label_id, repository, tag, **kwargs):
        client = self._get_client(**kwargs)
        return client.repositories_repo_name_tags_tag_labels_post(repository,
            tag, swagger_client.Label(id=int(label_id)))

    def delete_label(self, label_id, **kwargs):
        client = self._get_client(**kwargs)
        return client.labels_id_delete_with_http_info(int(label_id))