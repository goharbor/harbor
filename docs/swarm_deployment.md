**IMPORTANT** This deployment method is experimental, Docker Swarm experience is still required for volumes and container placement constraints!

## Docker Swarm

This document describes how to deploy **harbor** with **Docker Swarm**.
* You should have domain knowledge about **Docker Swarm** (distributed volumes, placement constraints)
* Tested with `master` branch

### Configuration

Change your settings in `harbor.cfg` as usual. For the deployment with swarm the configuration has to be prepared using the `./prepare` script and the following parameters:

```
./prepare --experimental-swarm
```

### clair

When using **clair**  comment the according lines in `harbor-stack.yml` and enable **clair**:

```
./prepare --with-clair --experimental-swarm
```


### SSL

#### nginx

When using SSL with the supplied **nginx** uncomment the according lines in `harbor-stack.yml` and change the settings in `harbor.cfg`:

```
protocol=https
ssl_cert = ./data/cert/server.crt
ssl_cert_key = ./data/cert/server.key
```

Then place your cert and key files in `./data/cert/`.

#### traefik
If you are using SSL with **traefik**, you have to change the realm property in `./common/config/registry/config.yml`: Go to `auth` -> `token` -> `realm` and replace "http" with "https". See: https://github.com/goharbor/harbor/issues/1097

Afterwards remove the `nginx`-service section in `./harbor-stack.yml` and add your traefik labels to the `portal`-service


### Deploy stack
**IMPORTANT** The stack configuration in `harbor-stack.yml` has to be modified manually to match your swarm environment:

* **Take care of the volume locations!**
* **Modify the placement constraints!**

Afterwards deploy **harbor** to you swarm:

```
docker stack deploy -c ./harbor-stack.yml harbor
```
