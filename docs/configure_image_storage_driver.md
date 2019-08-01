# Config Harbor image storage driver

Harbor is able to compute capacity of the volume used to store images.
By default it check the capacity of the `/data` volume.

## Configure OpenStack Swift driver

[OpenStack Swift](https://docs.openstack.org/swift/latest/) is a backend for Docker registry.
Harbor is able to get capacity of this remote storage with few environment variables.

To enable Swift driver, export following variables in the Harbor Core deployment.

```bash
# Tell harbor you use Swift driver
IMAGE_STORE_SWIFT_DRIVER=swift

# Get the `openrc` file from your Cloud provider
OS_AUTH_URL=https://auth.cloud.ovh.net/v3/
OS_DOMAIN_NAME=Default
OS_TENANT_ID=o123zq567vuer753oquz6broq6zi6u6b
OS_REGION_NAME=GRA5
OS_IDENTITY_API_VERSION=3
OS_TENANT_NAME=th3t3n4ntnAm3
OS_PASSWORD=***
OS_USERNAME=th3u5er

# Add a custom variable for the name of the Swift container you use
OS_CONTAINER_NAME=my-container
```
