#!/usr/bin/env python
# -*- coding: utf-8 -*-


from __future__ import print_function
import argparse
import os
import sys
import utils
import importlib
import glob
import shutil
import sys

def main():
    target_version = '1.8.0'
    parser = argparse.ArgumentParser(description='migrator of harbor.cfg') 
    parser.add_argument('--input', '-i', action="store", dest='input_path', required=True, help='The path to the old harbor.cfg that provides input value, this required value')
    parser.add_argument('--output','-o', action="store", dest='output_path', required=False, help='The path of the migrated harbor.cfg, if not set the input file will be overwritten')
    parser.add_argument('--target', action="store", dest='target_version', help='The target version that the harbor.cfg will be migrated to.')
    args = parser.parse_args()
    if args.output_path is None:
        args.output_path = args.input_path
    if args.target_version is not None:
        target_version = args.target_version
    input_version =  utils.get_conf_version(args.input_path)
    curr_dir = os.path.dirname(__file__)
    chain = []
    if input_version == target_version:
        print ("Version of input harbor.cfg is identical to target %s, no need to upgrade" % input_version)
        sys.exit(0)
    if not search(curr_dir, input_version, target_version, chain):
        print ("No migrator for version: %s" % input_version)
        sys.exit(1)
    else:
        print ("input version: %s, migrator chain: %s" % (input_version, chain))
    curr_input_path = args.input_path
    for c in chain:
    #TODO: more real-world testing needed for chained-migration.
        m = importlib.import_module(to_module_path(c))
        curr_output_path = "harbor.cfg.%s.tmp" % c
        print("migrating to version %s" % c)
        m.migrate(curr_input_path, curr_output_path)
        curr_input_path = curr_output_path
    shutil.copy(curr_output_path, args.output_path)
    print("Written new values to %s" % args.output_path)
    for tmp_f in glob.glob("harbor.cfg.*.tmp"):
        os.remove(tmp_f)

def to_module_path(ver):
    return "migrator_%s" % ver.replace(".","_")

def search(basedir, input_ver, target_ver, l):
    module = to_module_path(target_ver)
    if os.path.isdir(os.path.join(basedir, module)):
        m = importlib.import_module(module)
        if input_ver in m.acceptable_versions:
            l.append(target_ver)
            return True
        for v in m.acceptable_versions:
            if search(basedir, input_ver, v, l):
                l.append(target_ver)
                return True
    return False

if __name__ == "__main__":
    main()
