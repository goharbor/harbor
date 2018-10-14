#!/usr/bin/env python
# -*- coding:utf-8 -*-
# bug-report: feilengcui008@gmail.com

""" cli tool """

import argparse
import sys
import json
from datetime import datetime
from registry import RegistryApi
from registry import RegistryException


class ApiProxy(object):
    """ user RegistryApi """
    def __init__(self, registry, args):
        self.registry = registry
        self.args = args
        self.callbacks = dict()
        self.register_callback("repo", "list", self.list_repo)
        self.register_callback("tag", "list", self.list_tag)
        self.register_callback("tag", "delete", self.delete_tag)
        self.register_callback("manifest", "list", self.list_manifest)
        self.register_callback("manifest", "delete", self.delete_manifest)
        self.register_callback("manifest", "get", self.get_manifest)
        self.register_callback("tag", "remove_list", self.remove_tags_by_date)
        self.register_callback("tag", "list_date", self.list_tag_date)

    def register_callback(self, target, action, func):
        """ register real actions """
        if not target in self.callbacks.keys():
            self.callbacks[target] = {action: func}
            return
        self.callbacks[target][action] = func

    def execute(self, target, action):
        """ execute """
        print json.dumps(self.callbacks[target][action](), indent=4, sort_keys=True, default=datetime_converter)

    def list_repo(self):
        """ list repo """
        return self.registry.getRepositoryList(self.args.num)

    def list_tag(self):
        """ list tag """
        return self.registry.getTagList(self.args.repo)

    def delete_tag(self):
        """ delete tag """
        (_, ref) = self.registry.existManifest(self.args.repo, self.args.tag)
        if ref is not None:
            return self.registry.deleteManifest(self.args.repo, ref)
        return False

    def list_manifest(self):
        """ list manifest """
        tags = self.registry.getTagList(self.args.repo)["tags"]
        manifests = list()
        if tags is None:
            return None
        for i in tags:
            content = self.registry.getManifestWithConf(self.args.repo, i)
            manifests.append({i: content})
        return manifests

    def list_tag_date(self, tags=None):
        """ sort tags """
        manifests = list()
        if tags is None:
            tags = self.registry.getTagList(self.args.repo)["tags"]
            if tags is None:
                return None
        tags_nums = len(tags)
        for i in range(tags_nums):
            if i % 10 == 0:
                print "progress: {}%".format(100 * i / tags_nums)
            try:
                content = self.registry.getManifestWithConf(self.args.repo, tags[i])['configContent']['created']
                tag_time = datetime.strptime(content.split('T')[0], "%Y-%m-%d")
                manifests.append({'name': tags[i], 'date': tag_time})
            except RegistryException:
                pass
        return manifests

    def remove_tags_by_date(self):
        self.args.num = None
        repos = self.list_repo()
        for repo in repos['repositories']:
            print "repository: %s" % repo
            tags = self.registry.getTagList(repo)["tags"]
            tags_nums = len(tags)
            print "number of tags in repo: %s" % str(tags_nums)
            if tags_nums > int(self.args.number):
                self.args.repo = repo
                manifests = self.list_tag_date(tags)
                sort_tags = sorted(manifests, key=lambda manifests: manifests['date'])
                for i in range(tags_nums - int(self.args.number)):
                    if not self.args.dryrun:
                        self.args.tag = sort_tags[i].get('name')
                        self.delete_tag()
                    else:
                        print "dry-run mode (no tags will be removed)"
                    print "tag %s removed" % sort_tags[i].get('name')
        return 'tags successfully removed'

    def delete_manifest(self):
        """ delete manifest """
        return self.registry.deleteManifest(self.args.repo, self.args.ref)

    def get_manifest(self):
        """ get manifest """
        return self.registry.getManifestWithConf(self.args.repo, self.args.tag)


def datetime_converter(o):
    if isinstance(o, datetime):
        return o.__str__()


# since just a script tool, we do not construct whole target->action->args
# structure with oo abstractions which has more flexibility, just register
# parser directly
def get_parser():
    """ return a parser """
    parser = argparse.ArgumentParser("cli")

    parser.add_argument('--username', action='store', required=True, help='username')
    parser.add_argument('--password', action='store', required=True, help='password')
    parser.add_argument('--registry_endpoint', action='store', required=True,
            help='registry endpoint')

    subparsers = parser.add_subparsers(dest='target', help='target to operate on')

    # repo target
    repo_target_parser = subparsers.add_parser('repo', help='target repository')
    repo_target_subparsers = repo_target_parser.add_subparsers(dest='action',
            help='repository subcommand')
    repo_cmd_parser = repo_target_subparsers.add_parser('list', help='list repositories')
    repo_cmd_parser.add_argument('--num', action='store', required=False, default=None,
            help='the number of data to return')

    # tag target
    tag_target_parser = subparsers.add_parser('tag', help='target tag')
    tag_target_subparsers = tag_target_parser.add_subparsers(dest='action',
            help='tag subcommand')
    tag_list_parser = tag_target_subparsers.add_parser('list', help='list tags')
    tag_list_parser.add_argument('--repo', action='store', required=True, help='list tags')
    tag_delete_parser = tag_target_subparsers.add_parser('delete', help='delete tag')
    tag_delete_parser.add_argument('--repo', action='store', required=True, help='delete tags')
    tag_delete_parser.add_argument('--tag', action='store', required=True,
                                   help='tag reference')
    tag_list_delete_parser = tag_target_subparsers.add_parser('remove_list', help='remove list tags')
    tag_list_delete_parser.add_argument('--number', action='store', required=True, help='remove list tags')
    tag_list_delete_parser.add_argument('--dryrun', action='store_true', required=False, help='remove list tags')

    tag_list_date_parser = tag_target_subparsers.add_parser('list_date', help='list tags by date created')
    tag_list_date_parser.add_argument('--repo', action='store', required=True, help='list tags by date created')

    # manifest target
    manifest_target_parser = subparsers.add_parser('manifest', help='target manifest')
    manifest_target_subparsers = manifest_target_parser.add_subparsers(dest='action',
            help='manifest subcommand')
    manifest_list_parser = manifest_target_subparsers.add_parser('list', help='list manifests')
    manifest_list_parser.add_argument('--repo', action='store', required=True,
            help='list manifests')
    manifest_delete_parser = manifest_target_subparsers.add_parser('delete', help='delete manifest')
    manifest_delete_parser.add_argument('--repo', action='store', required=True,
            help='delete manifest')
    manifest_delete_parser.add_argument('--ref', action='store', required=True,
            help='manifest reference')
    manifest_get_parser = manifest_target_subparsers.add_parser('get', help='get manifest content')
    manifest_get_parser.add_argument('--repo', action='store', required=True, help='delete tags')
    manifest_get_parser.add_argument('--tag', action='store', required=True,
            help='manifest reference')
    
    return parser


def main():
    """ main entrance """
    parser = get_parser()
    options = parser.parse_args(sys.argv[1:])
    registry = RegistryApi(options.username, options.password, options.registry_endpoint)
    proxy = ApiProxy(registry, options)
    proxy.execute(options.target, options.action)


if __name__ == '__main__':
    main()
