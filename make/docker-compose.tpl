version: '2'
services:
  log:
    image: goharbor/harbor-log:__version__
    container_name: harbor-log 
    restart: always
    dns_search: .
    volumes:
      - /var/log/harbor/:/var/log/docker/:z
      - ./common/config/log/:/etc/logrotate.d/:z
    ports:
      - 127.0.0.1:1514:10514
    networks:
      - harbor
  registry:
    image: goharbor/registry-photon:__reg_version__
    container_name: registry
    restart: always
    volumes:
      - /data/registry:/storage:z
      - ./common/config/registry/:/etc/registry/:z
    networks:
      - harbor
    dns_search: .
    depends_on:
      - log
    logging:
      driver: "syslog"
      options:  
        syslog-address: "tcp://127.0.0.1:1514"
        tag: "registry"
  registryctl:
    image: goharbor/harbor-registryctl:__version__
    container_name: registryctl
    env_file:
      - ./common/config/registryctl/env
    restart: always
    volumes:
      - /data/registry:/storage:z
      - ./common/config/registry/:/etc/registry/:z
      - ./common/config/registryctl/config.yml:/etc/registryctl/config.yml:z
    networks:
      - harbor
    dns_search: .
    depends_on:
      - log
    logging:
      driver: "syslog"
      options:  
        syslog-address: "tcp://127.0.0.1:1514"
        tag: "registryctl"
  postgresql:
    image: goharbor/harbor-db:__version__
    container_name: harbor-db
    restart: always
    volumes:
      - /data/database:/var/lib/postgresql/data:z
    networks:
      - harbor
    dns_search: .
    env_file:
      - ./common/config/db/env
    depends_on:
      - log
    logging:
      driver: "syslog"
      options:  
        syslog-address: "tcp://127.0.0.1:1514"
        tag: "postgresql"
  adminserver:
    image: goharbor/harbor-adminserver:__version__
    container_name: harbor-adminserver
    env_file:
      - ./common/config/adminserver/env
    restart: always
    volumes:
      - /data/config/:/etc/adminserver/config/:z
      - /data/secretkey:/etc/adminserver/key:z
      - /data/:/data/:z
    networks:
      - harbor
    dns_search: .
    depends_on:
      - log
    logging:
      driver: "syslog"
      options:  
        syslog-address: "tcp://127.0.0.1:1514"
        tag: "adminserver"
  core:
    image: goharbor/harbor-core:__version__
    container_name: harbor-core
    env_file:
      - ./common/config/core/env
    restart: always
    volumes:
      - ./common/config/core/app.conf:/etc/core/app.conf:z
      - ./common/config/core/private_key.pem:/etc/core/private_key.pem:z
      - ./common/config/core/certificates/:/etc/core/certificates/:z
      - /data/secretkey:/etc/core/key:z
      - /data/ca_download/:/etc/core/ca/:z
      - /data/psc/:/etc/core/token/:z
      - /data/:/data/:z
    networks:
      - harbor
    dns_search: .
    depends_on:
      - log
      - adminserver
      - registry
    logging:
      driver: "syslog"
      options:  
        syslog-address: "tcp://127.0.0.1:1514"
        tag: "core"
  portal:
    image: goharbor/harbor-portal:__version__
    container_name: harbor-portal
    restart: always
    networks:
      - harbor
    dns_search: .
    depends_on:
      - log
      - core
    logging:
      driver: "syslog"
      options:
        syslog-address: "tcp://127.0.0.1:1514"
        tag: "portal"

  jobservice:
    image: goharbor/harbor-jobservice:__version__
    container_name: harbor-jobservice
    env_file:
      - ./common/config/jobservice/env
    restart: always
    volumes:
      - /data/job_logs:/var/log/jobs:z
      - ./common/config/jobservice/config.yml:/etc/jobservice/config.yml:z
    networks:
      - harbor
    dns_search: .
    depends_on:
      - redis
      - core
      - adminserver
    logging:
      driver: "syslog"
      options:
        syslog-address: "tcp://127.0.0.1:1514"
        tag: "jobservice"
  redis:
    image: goharbor/redis-photon:__redis_version__
    container_name: redis
    restart: always
    volumes:
      - /data/redis:/var/lib/redis
    networks:
      - harbor
    dns_search: .
    depends_on:
      - log
    logging:
      driver: "syslog"
      options:
        syslog-address: "tcp://127.0.0.1:1514"
        tag: "redis"
  proxy:
    image: goharbor/nginx-photon:__version__
    container_name: nginx
    restart: always
    volumes:
      - ./common/config/nginx:/etc/nginx:z
    networks:
      - harbor
    dns_search: .
    ports:
      - 80:80
      - 443:443
      - 4443:4443
    depends_on:
      - postgresql
      - registry
      - core
      - portal
      - log
    logging:
      driver: "syslog"
      options:  
        syslog-address: "tcp://127.0.0.1:1514"
        tag: "proxy"
networks:
  harbor:
    external: false

