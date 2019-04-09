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

def get_storage_provider_info(provider_name, provider_config):
    provider_config = provider_config.strip('" ')
    if not provider_config.strip(" "):
        return ''

    storage_provider_cfg_map = {}
    for k_v in provider_config.split(","):
        if k_v > 0:
            kvs = k_v.split(": ") # add space suffix to avoid existing ":" in the value
            if len(kvs) == 2:
                #key must not be empty
                if kvs[0].strip() != "":
                    storage_provider_cfg_map[kvs[0].strip()] = kvs[1].strip()

    # generate storage configuration section in yaml format

    storage_provider_conf_list = [provider_name + ':']
    for config in storage_provider_cfg_map.items():
        storage_provider_conf_list.append('{}: {}'.format(*config))
    storage_provider_info = ('\n' + ' ' * 4).join(storage_provider_conf_list)
    return storage_provider_info
