# -*- coding: utf-8 -*-
import requests

def call(server, project_name, repo_name, digest, artifactType=None, **kwargs):
    url=None
    auth = (kwargs.get("username"), kwargs.get("password"))
    if artifactType:
        artifactType = artifactType.replace("+", "%2B")
        url="https://{}/v2/{}/{}/referrers/{}?artifactType={}".format(server, project_name, repo_name, digest, artifactType)
    else:
        url="https://{}/v2/{}/{}/referrers/{}".format(server, project_name, repo_name, digest)
    return requests.get(url, auth=auth, verify=False)
