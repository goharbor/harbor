<!--[metadata]>
+++
title = "Testing an insecure registry"
description = "Deploying a Registry in an insecure fashion"
keywords = ["registry, on-prem, images, tags, repository, distribution, insecure"]
+++
<![end-metadata]-->

# Insecure Registry

While it's highly recommended to secure your registry using a TLS certificate issued by a known CA, you may alternatively decide to use self-signed certificates, or even use your registry over plain http.

You have to understand the downsides in doing so, and the extra burden in configuration.

## Deploying a plain HTTP registry

> **Warning**: it's not possible to use an insecure registry with basic authentication

This basically tells Docker to entirely disregard security for your registry.

1. edit the file `/etc/default/docker` so that there is a line that reads: `DOCKER_OPTS="--insecure-registry myregistrydomain.com:5000"` (or add that to existing `DOCKER_OPTS`)
2. restart your Docker daemon: on ubuntu, this is usually `service docker stop && service docker start`

**Pros:**

 - relatively easy to configure
 
**Cons:**
 
 - this is **very** insecure: you are basically exposing yourself to trivial MITM, and this solution should only be used for isolated testing or in a tightly controlled, air-gapped environment
 - you have to configure every docker daemon that wants to access your registry 
  
## Using self-signed certificates

> **Warning**: using this along with basic authentication requires to **also** trust the certificate into the OS cert store for some versions of docker (see below)

Generate your own certificate:

    mkdir -p certs && openssl req \
      -newkey rsa:4096 -nodes -sha256 -keyout certs/domain.key \
      -x509 -days 365 -out certs/domain.crt

Be sure to use the name `myregistrydomain.com` as a CN.

Use the result to [start your registry with TLS enabled](https://github.com/docker/distribution/blob/master/docs/deploying.md#get-a-certificate)

Then you have to instruct every docker daemon to trust that certificate. This is done by copying the `domain.crt` file to `/etc/docker/certs.d/myregistrydomain.com:5000/ca.crt`.

Don't forget to restart docker after doing so.

**Pros:**

 - more secure than the insecure registry solution

**Cons:**

 - you have to configure every docker daemon that wants to access your registry

## Failing...

Failing to configure docker and trying to pull from a registry that is not using TLS will result in the following message:

```
FATA[0000] Error response from daemon: v1 ping attempt failed with error:
Get https://myregistrydomain.com:5000/v1/_ping: tls: oversized record received with length 20527. 
If this private registry supports only HTTP or HTTPS with an unknown CA certificate,please add 
`--insecure-registry myregistrydomain.com:5000` to the daemon's arguments.
In the case of HTTPS, if you have access to the registry's CA certificate, no need for the flag;
simply place the CA certificate at /etc/docker/certs.d/myregistrydomain.com:5000/ca.crt
```

## Docker still complains about the certificate when using authentication?

When using authentication, some versions of docker also require you to trust the certificate at the OS level.

Usually, on Ubuntu this is done with:

    cp certs/domain.crt /usr/local/share/ca-certificates/myregistrydomain.com.crt
    update-ca-certificates

... and on Red Hat (and its derivatives) with:

    cp certs/domain.crt /etc/pki/ca-trust/source/anchors/myregistrydomain.com.crt
    update-ca-trust

... On some distributions, e.g. Oracle Linux 6, the Shared System Certificates feature needs to be manually enabled:

    update-ca-trust enable

Now restart docker (`service docker stop && service docker start`, or any other way you use to restart docker).
