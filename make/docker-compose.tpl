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
    networks:
      - harbor
  registry:
    image: library/registry:2.5.1
    container_name: registry
    restart: always
    volumes:
      - /data/registry:/storage
      - ./common/config/registry/:/etc/registry/
    networks:
      - harbor
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
    networks:
      - harbor
    env_file:
      - ./common/config/db/env
    depends_on:
      - log
    logging:
      driver: "syslog"
      options:  
        syslog-address: "tcp://127.0.0.1:1514"
        tag: "mysql"
  etcd:
    container_name: etcd0
    image: quay.io/coreos/etcd:v3.0.15
    command: >
       /usr/local/bin/etcd
       -name etcd0
       -advertise-client-urls http://${HOSTIP}:2379,http://${HOSTIP}:4001
       -listen-client-urls http://0.0.0.0:2379,http://0.0.0.0:4001
       -initial-advertise-peer-urls http://${HOSTIP}:2380
       -listen-peer-urls http://0.0.0.0:2380
       -initial-cluster-token etcd-cluster-1
       -initial-cluster etcd0=http://${HOSTIP}:2380
       -initial-cluster-state new
    volumes:
      - /data/certs/:/etc/ssl/certs
    ports:
     - "2380:2380"
     - "2379:2379"
     - "4001:4001"

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
    networks:
      - harbor
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
    networks:
      - harbor
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
    networks:
      - harbor
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
networks:
  harbor:
    external: false

