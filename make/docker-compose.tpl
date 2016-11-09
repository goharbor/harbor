version: '2'
services:
  log:
    image: vmware/harbor-log
    container_name: harbor-log 
    restart: always
    volumes:
      - /var/log/harbor/:/var/log/docker/
    ports:
      - 1514:514
  registry:
    image: library/registry:2.5.0
    container_name: registry
    restart: always
    volumes:
      - /data/registry:/storage
      - ./common/config/registry/:/etc/registry/
    environment:
      - GODEBUG=netdns=cgo
    command:
      ["serve", "/etc/registry/config.yml"]
    depends_on:
      - log
    logging:
      driver: "syslog"
      options:  
        syslog-address: "tcp://127.0.0.1:1514"
        tag: "registry"
  mysql:
    image: vmware/harbor-db
    container_name: harbor-db
    restart: always
    volumes:
      - /data/database:/var/lib/mysql
    env_file:
      - ./common/config/db/env
    depends_on:
      - log
    logging:
      driver: "syslog"
      options:  
        syslog-address: "tcp://127.0.0.1:1514"
        tag: "mysql"
  ui:
    image: vmware/harbor-ui
    container_name: harbor-ui
    env_file:
      - ./common/config/ui/env
    restart: always
    volumes:
      - ./common/config/ui/app.conf:/etc/ui/app.conf
      - ./common/config/ui/private_key.pem:/etc/ui/private_key.pem
      - /data:/harbor_storage
    depends_on:
      - log
    logging:
      driver: "syslog"
      options:  
        syslog-address: "tcp://127.0.0.1:1514"
        tag: "ui"
  jobservice:
    image: vmware/harbor-jobservice
    container_name: harbor-jobservice
    env_file:
      - ./common/config/jobservice/env
    restart: always
    volumes:
      - /data/job_logs:/var/log/jobs
      - ./common/config/jobservice/app.conf:/etc/jobservice/app.conf
    depends_on:
      - ui
    logging:
      driver: "syslog"
      options:  
        syslog-address: "tcp://127.0.0.1:1514"
        tag: "jobservice"
  proxy:
    image: nginx:1.11.5
    container_name: nginx
    restart: always
    volumes:
      - ./common/config/nginx:/etc/nginx
    ports:
      - 80:80
      - 443:443
    depends_on:
      - mysql
      - registry
      - ui
      - log
    logging:
      driver: "syslog"
      options:  
        syslog-address: "tcp://127.0.0.1:1514"
        tag: "proxy"
