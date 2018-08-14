# -*- coding: utf-8 -*-

import sys
import os
# It's a temporary solution to workaround the "No module named xxx" error 
# when importing the Harbor.py as a robot library
# Line 8-11 should be removed when the root cause found 
with open("/usr/local/lib/python2.7/dist-packages/easy-install.pth") as f:
    for line in  f.readlines():
        sys.path.append(os.path.join("/usr/local/lib/python2.7/dist-packages/", line[2:].rstrip()))
sys.path.append(os.path.join(os.path.dirname(__file__),"../../../../harborclient"))
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