#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
This is a script to get npm token
"""

import os
import json
import httplib


def main():
    """
    get token from npm
    """
    username = os.getenv("NPM_USERNAME")
    password = os.getenv("NPM_PASSWORD")

    headers = {'Accept': 'application/json', 'Content-Type': 'application/json'}
    auth = {'name': username, 'password': password}
    data = json.dumps(auth)
    conn = httplib.HTTPSConnection("registry.npmjs.org")
    conn.request('PUT', '/-/user/org.couchdb.user:{name}'.format(**auth), data, headers)
    res = conn.getresponse()

    if int(res.status) / 100 != 2:
        raise Exception("npm response not 2XX status")
    print json.loads(res.read())['token']

if __name__ == '__main__':
    main()
