# Troubleshooting Harbor Installation

1. When Harbor does not work properly, run the below commands to find out if all containers of Harbor are in **UP** status:
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

If a container is not in **UP** state, check the log file of that container in directory `/var/log/harbor`. For example, if the container `harbor-core` is not running, you should look at the log file `core.log`.

2.When setting up Harbor behind an nginx proxy or elastic load balancing, look for the line below, in `common/config/nginx/nginx.conf` and remove it from the sections if the proxy already has similar settings: `location /`, `location /v2/` and `location /service/`.

``` sh
proxy_set_header X-Forwarded-Proto $scheme;
```

and re-deploy Harbor refer to the previous section "Managing Harbor's lifecycle".
