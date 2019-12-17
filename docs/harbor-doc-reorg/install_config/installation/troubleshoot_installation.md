# Troubleshooting Harbor Installation

## Harbor Doesn't Start or Functions Incorrectly

Harbor Doesn't Start or Functions Incorrectly

When Harbor does not function correctly, run the following commands to find out if all of Harbor's containers in **UP** status:
```
    $ sudo docker-compose ps
        Name                     Command               State                    Ports
  -----------------------------------------------------------------------------------------------------------------------------
  harbor-core         /harbor/start.sh                 Up
  harbor-db           /entrypoint.sh postgres          Up      5432/tcp
  harbor-jobservice   /harbor/start.sh                 Up
  harbor-log          /bin/sh -c /usr/local/bin/ ...   Up      127.0.0.1:1514->10514/tcp
  harbor-portal       nginx -g daemon off;             Up      80/tcp
  nginx               nginx -g daemon off;             Up      0.0.0.0:443->443/tcp, 0.0.0.0:4443->4443/tcp, 0.0.0.0:80->80/tcp
  redis               docker-entrypoint.sh redis ...   Up      6379/tcp
  registry            /entrypoint.sh /etc/regist ...   Up      5000/tcp
  registryctl         /harbor/start.sh                 Up
```

If a container is not in the `Up` state, check the log file for that container in `/var/log/harbor`. For example, if the `harbor-core` container is not running, look at the `core.log` log file.

### Using nginx or Load Balancing

When setting up Harbor behind an `nginx` proxy or elastic load balancing, look for the following line in `common/config/nginx/nginx.conf` and, if the proxy already has similar settings, remove it from the sections `location /`, `location /v2/` and `location /service/`.

``` sh
proxy_set_header X-Forwarded-Proto $scheme;
```

Then re-deploy Harbor per the instructions in "Managing Harbor Lifecycle.

[Back to table of contents](../../_index.md)