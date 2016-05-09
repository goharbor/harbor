<!--[metadata]>
+++
title = "Ceph RADOS storage driver"
description = "Explains how to use the Ceph RADOS storage driver"
keywords = ["registry, service, driver, images, storage, ceph,  rados"]
+++
<![end-metadata]-->


# Ceph RADOS storage driver

An implementation of the `storagedriver.StorageDriver` interface which uses
[Ceph RADOS Object Storage][rados] for storage backend.

## Parameters


<table>
  <tr>
    <th>Parameter</th>
    <th>Required</th>
    <th>Description</th>
  </tr>
  <tr>
    <td>
      <code>poolname</code>
    </td>
    <td>
      yes
    </td>
    <td>
      Ceph pool name.
    </td>
  </tr>
   <tr>
    <td>
      <code>username</code>
    </td>
    <td>
      no
    </td>
    <td>
      Ceph cluster user to connect as (i.e. admin, not client.admin).
    </td>
  </tr>
   <tr>
    <td>
      <code>chunksize</code>
    </td>
    <td>
      no
    </td>
    <td>
      Size of the written RADOS objects. Default value is 4MB (4194304).
    </td>
  </tr>
</table>


The following parameters must be used to configure the storage driver
(case-sensitive):

* `poolname`: Name of the Ceph pool
* `username` *optional*: The user to connect as (i.e. admin, not client.admin)
* `chunksize` *optional*: Size of the written RADOS objects. Default value is
4MB (4194304).

This drivers loads the [Ceph client configuration][rados-config] from the
following regular paths (the first found is used):

* `$CEPH_CONF` (environment variable)
* `/etc/ceph/ceph.conf`
* `~/.ceph/config`
* `ceph.conf` (in the current working directory)

## Developing

To include this driver when building Docker Distribution, use the build tag
`include_rados`. Please see the [building documentation][building] for details.

[rados]: http://ceph.com/docs/master/rados/
[rados-config]: http://ceph.com/docs/master/rados/configuration/ceph-conf/
[building]: https://github.com/docker/distribution/blob/master/docs/building.md#optional-build-tags
