import os
import logging

from buildwebapi import api as buildapi
LOG = logging.getLogger(__name__)


def get_build_type(build_id):
    build = get_build(build_id)
    LOG.debug('%s is %s build', build_id, build.buildtype)
    return build.buildtype


def get_build_id_and_system(build_id):
    build_system = 'ob'
    if '-' in str(build_id):
        temp = build_id.split('-')
        build_id = temp[1]
        build_system = temp[0]
    return build_id, build_system


def get_ova_url(build_id):
    return get_url(build_id, '_OVF10.ova')


def get_url(build_id, deliverable_name):
    build = get_build(build_id)
    deliverables = buildapi.ListResource.by_url(build._deliverables_url)
    deliverable = [d for d in deliverables
                   if d.matches(path=deliverable_name)][0]
    LOG.debug('Download URL of %s is %s', build_id, deliverable._download_url)
    return deliverable._download_url


def get_product(build_id):
    build = get_build(build_id)
    LOG.debug('Product of %s is %s.', build_id, build.product)
    return build.product


def get_latest_build_url(branch, build_type, product='harbor_build'):
    build_id = get_latest_build_id(branch, build_type, product)
    print build_id
    return get_ova_url(build_id)


def get_latest_build_id(branch, build_type, product='harbor_build'):
    return buildapi.MetricResource.by_name('build',
                                           product=product,
                                           buildstate='succeeded',
                                           buildtype=build_type,
                                           branch=branch).get_max_id()


def get_build(build_id):
    build_id, build_system = get_build_id_and_system(build_id)
    return buildapi.ItemResource.by_id('build', int(build_id), build_system)


def get_build_version(build_id):
    build = get_build(build_id)
    LOG.debug('Version of %s is %s.', build_id, build.version)
    return build.version