#!/usr/bin/env python
# -*- coding: utf-8 -*-
import os
import yaml
import click
import importlib

from collections import deque

BASE_DIR = os.path.dirname(__file__)

def read_conf(path):
    with open(path) as f:
        try:
            d = yaml.safe_load(f)
        except Exception as e:
            click.echo("parse config file err, make sure your harbor config version is above 1.8.0", e)
            exit(-1)
    return d

def _to_module_path(ver):
    return "migration.versions.{}".format(ver.replace(".","_"))

def search(input_ver: str, target_ver: str) -> deque :
    """
    Search accept a input version and the target version.
    Returns the module of migrations in the upgrade path
    """
    upgrade_path, visited = deque(), set()
    while True:
        module_path = _to_module_path(target_ver)
        visited.add(target_ver)  # mark current version for loop finding
        if os.path.isdir(os.path.join(BASE_DIR, 'versions', target_ver.replace(".","_"))):
            module = importlib.import_module(module_path)
            if module.revision == input_ver:    # migration path found
                break
            elif module.down_revision is None:  # migration path not found
                raise Exception('no migration path found')
            else:
                upgrade_path.appendleft(module)
                target_ver = module.down_revision
                if target_ver in visited: # version visited before, loop found
                    raise Exception('find a loop caused by {} on migration path'.format(target_ver))
        else:
            raise Exception('{} not dir'.format(os.path.join(BASE_DIR, 'versions', target_ver.replace(".","_"))))
    return upgrade_path
