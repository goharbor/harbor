#!/usr/bin/env python
# -*- coding: utf-8 -*-
import sys
import os
import json
from string import Template

if sys.version_info[:3][0] == 2:
    import ConfigParser as ConfigParser
    import StringIO as StringIO

if sys.version_info[:3][0] == 3:
    import configparser as ConfigParser
    import io as StringIO

def read_conf(path):
    temp_section = "configuration"
    conf = StringIO.StringIO()
    conf.write("[%s]\n" % temp_section)
    conf.write(open(path).read())
    conf.seek(0, os.SEEK_SET)
    rcp = ConfigParser.RawConfigParser()
    rcp.readfp(conf)
    d = {}
    for op in rcp.options(temp_section):
        d[op] = rcp.get(temp_section, op)
    return d

def get_conf_version(path):
    d = read_conf(path)
#    print json.dumps(d,indent=4)
    if "_version" in d: # >=1.5.0
        return d["_version"]
    if not "clair_db_password" in d:
        return "unsupported"
    if "registry_storage_provider_name" in d:
        return "1.4.0"
    if "uaa_endpoint" in d:
        return "1.3.0"
    return "1.2.0"

def render(src, dest, **kw):
    t = Template(open(src, 'r').read())
    with open(dest, 'w') as f:
        f.write(t.substitute(**kw))
