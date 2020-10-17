from __future__ import absolute_import


import unittest
import numpy
import threading
from datetime import *
from time import sleep, ctime

import library.repository
import library.docker_api
from library.base import _assert_status_code
from testutils import ADMIN_CLIENT
from testutils import harbor_server

from testutils import TEARDOWN
from library.project import Project
from library.repository import Repository
from library.artifact import Artifact
from library.repository import push_image_to_project
from library.repository import pull_harbor_image
from library.repository import push_special_image_to_project
import argparse

def do_populate(name_index, repo_count):
    project= Project()
    artifact = Artifact()
    repo = Repository()
    url = ADMIN_CLIENT["endpoint"]
    ADMIN_CLIENT["password"] =  "qA5ZgV"


    #2. Create a new project(PA) by user(UA);
    project_name = "project"+str(name_index)
    if project.check_project_name_exist(name=project_name, **ADMIN_CLIENT) is not True:
        project.create_project(name=project_name, metadata = {"public": "false"}, **ADMIN_CLIENT)
        print("Create Project:", project_name)

    tag = 'latest'
    for repo_index in range(int(repo_count)):
        repo_name = "image"+ str(repo_index)
        if artifact.check_reference_exist(project_name, repo_name, tag, ignore_not_found=True, **ADMIN_CLIENT) is not True:
            push_special_image_to_project(project_name, harbor_server, ADMIN_CLIENT["username"], ADMIN_CLIENT["password"], repo_name, [tag], size=repo_index*30)
            print("Push Image:", repo_name)
            for tag_index in numpy.arange(1, 2, 0.1):
                artifact.create_tag(project_name, repo_name, tag, str(tag_index), ignore_conflict = True, **ADMIN_CLIENT)
                print("Add Tag:", str(tag_index))


def get_parser():
    """ return a parser """
    parser = argparse.ArgumentParser("populate")

    parser.add_argument('--project-count','-p', dest='project_count', required=False, default=100)
    parser.add_argument('--repo-count','-r', dest='repo_count', required=False, default=100)
    args = parser.parse_args()
    return (args.project_count, args.repo_count)

def main():
    """ main entrance """
    project_count, repo_count = get_parser()
    Threads = []
    for i in range(int(project_count)):
        t = threading.Thread(target=do_populate, args=(str(i), repo_count), name='T'+str(i))
        t.setDaemon(True)
        Threads.append(t)
    sleep(3)
    for t in Threads:
        t.start()
    for t in Threads:
        t.join()

    print('Job Finished:', ctime())

if __name__ == '__main__':
    main()

