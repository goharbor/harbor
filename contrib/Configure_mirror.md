# Configuring Harbor as a local registry mirror

Harbor runs as a local registry by default. It can also be configured as a registry mirror,
which caches downloaded images for subsequent use. Note that under this setup, the Harbor registry only acts as a mirror server and
no longer accepts image pushing requests. Edit `Deploy/templates/registry/config.yml` before executing `./prepare`, and append a `proxy` section as follows:

```
proxy:
  remoteurl: https://registry-1.docker.io
```
In order to access private images on the Docker Hub, a username and a password can be supplied:

```
proxy:
  remoteurl: https://registry-1.docker.io
  username: [username]
  password: [password]
```
You will need to pass the `--registry-mirror` option to your Docker daemon on startup:

```
docker --registry-mirror=https://<my-docker-mirror-host> daemon
```
For example, if your mirror is serving on `http://reg.yourdomain.com`, you would run:

```
docker --registry-mirror=https://reg.yourdomain.com daemon
```

Refer to the [Registry as a pull through cache](https://docs.docker.com/registry/recipes/mirror/) for detailed information.
