#!/usr/bin/env python
# -*- coding: utf-8 -*-
import sys
import os
import yaml
import click
import importlib
from string import Template

def read_conf(path):
    with open(path) as f:
        try:
            d = yaml.safe_load(f)
        except Exception as e:
            click.echo("parse config file err, make sure your harbor config version is above 1.8.0", e)
            exit(-1)
    return d

def to_module_path(ver):
    return "migration.versions.{}".format(ver.replace(".","_"))

def search(input_ver: str, target_ver: str) -> list:
    """
    Search accept a input version and the target version.
    Returns the module of migrations in the upgrade path
    """
    def helper():
        nonlocal basedir
        nonlocal cur_target
        while True:
            module_path = to_module_path(cur_target)
            if os.path.isdir(os.path.join(basedir, 'versions', cur_target.replace(".","_"))):
                module = importlib.import_module(module_path)
                yield module
                if module.revision == input_ver:
                    return
                elif module.down_revision is not None:
                    cur_target = module.down_revision
                else:
                    return
            else:
                click.echo('{} not dir'.format(os.path.join(basedir, 'versions', cur_target.replace(".","_"))))
                return
    basedir = os.path.dirname(__file__)
    cur_target = target_ver
    upgrade_path = list(helper())
    upgrade_path.reverse()
    if upgrade_path and upgrade_path[0].revision == input_ver and upgrade_path[-1].revision == target_ver:
        return upgrade_path[1:]
    else:
        return []
