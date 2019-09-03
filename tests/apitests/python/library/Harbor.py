# -*- coding: utf-8 -*-
import site
reload(site)
import project
import label
import registry
import replication
import repository
import swagger_client

class Harbor(project.Project, label.Label,
    registry.Registry, replication.Replication,
    repository.Repository):
    pass