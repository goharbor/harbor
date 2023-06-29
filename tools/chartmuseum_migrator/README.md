# Chartmuseum Migrator

[Harbor](https://github.com/goharbor/harbor) supports two different ways to store the Helms charts data:
    1. stored in Harbor registry storage directly via OCI API.
    2. stored in Harbor hosted chartmuseum backend via chartmuseum's API.

From Harbor 2.6, [Chartmuseum](https://github.com/helm/chartmuseum) is deprecated and is removed from Harbor 2.8.

Chartmuseum Migrator Tool purpose is to migrate Helm charts from Harbor Chartmuseum to Helm OCI registry.
It copies Helm charts but don't delete them from Chartmuseum.

## Requirements

- Docker

## Build

```bash
docker build -t goharbor/chartmuseum-migrator .
```

## Usage

```bash
docker run -ti --rm goharbor/chartmuseum-migrator --url {{harbor_hostname}} --username {{harbor_user}} --password {{harbor_password}}
```
