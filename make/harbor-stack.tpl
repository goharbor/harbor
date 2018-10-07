version: '3.4'
services:
  registry:
    image: goharbor/registry-photon:__reg_version__
    volumes:
      - registry:/storage:z
    secrets:
      - source: registry_root_crt
        target: /etc/registry/root.crt
    configs:
      - source: registry_config
        target: /etc/registry/config.yml
    networks:
      - harbor
    environment:
      - GODEBUG=netdns=cgo
    deploy:
      replicas: 1
      placement:
        constraints:
          - node.role == manager

  registryctl:
    image: goharbor/harbor-registryctl:__version__
    env_file:
      - ./common/config/registryctl/env
    volumes:
      - registry:/storage:z
    secrets:
      - source: registry_root_crt
        target: /etc/registry/root.crt
    configs:
      - source: registry_config
        target: /etc/registry/config.yml
      - source: registryctl_config
        target: /etc/registryctl/config.yml
    networks:
      - harbor
    environment:
      - GODEBUG=netdns=cgo
    deploy:
      replicas: 1
      placement:
        constraints:
          - node.role == manager

  postgresql:
    image: goharbor/harbor-db:__version__
    volumes:
      - database:/var/lib/postgresql/data:z
    networks:
      - harbor
    environment:
      - ./common/config/db/env
    deploy:
      replicas: 1
      placement:
        constraints:
          - node.role == manager

  adminserver:
    image: goharbor/harbor-adminserver:__version__
    env_file:
      - ./common/config/adminserver/env
    volumes:
      - admin_config:/etc/adminserver/config:z
      - admin_data:/data/:z
    secrets:
      - source: secretkey
        target: /etc/adminserver/key
    networks:
      - harbor
    deploy:
      replicas: 1
      placement:
        constraints:
          - node.role == manager

  portal:
    image: goharbor/harbor-portal:__version__
    networks:
      - harbor
    deploy:
      replicas: 1
      placement:
        constraints:
          - node.role == manager
  core:
    image: goharbor/harbor-core:__version__
    env_file:
      - ./common/config/core/env
    volumes:
      #- ./common/config/ui/certificates/:/etc/ui/certificates/:z
      - core_ca_download:/etc/core/ca/:z
      - core_psc:/etc/core/token/:z
    secrets:
      - source: secretkey
        target: /etc/core/key
      - source: core_privatekey
        target: /etc/core/private_key.pem
    # use more secrets for the certificates in /etc/ui/certificates instead of volumes
    configs:
      - source: core_config
        target: /etc/core/app.conf
    networks:
      - harbor
    deploy:
      replicas: 1
      placement:
        constraints:
          - node.role == manager

  jobservice:
    image: goharbor/harbor-jobservice:__version__
    env_file:
      - ./common/config/jobservice/env
    volumes:
      - jobs:/var/log/jobs:z
    configs:
      - source: jobservice_config
        target: /etc/jobservice/config.yml
    networks:
      - harbor
    deploy:
      replicas: 1
      placement:
        constraints:
          - node.role == manager

  redis:
    image: goharbor/redis-photon:__redis_version__
    volumes:
      - redis:/var/lib/redis
    networks:
      - harbor
    deploy:
      replicas: 1
      placement:
        constraints:
          - node.role == manager

# Uncomment when using clair
#  clair:
#    image: goharbor/clair-photon:__clair_version__
#    deploy:
#      replicas: 1
#      constraints:
#        - node.role == manager
#      resources:
#        limits:
#          cpus: '0.50'
#    networks:
#      - harbor
#    configs:
#      - source: clair_config
#        target: /etc/clair/config.yaml
#    env_file:
#      ./common/config/clair/clair_env

  proxy:
    image: goharbor/nginx-photon:__version__
    configs:
      - source: nginx_config
        target: /etc/nginx/nginx.conf
    deploy:
      replicas: 1
      placement:
        constraints:
          - node.role == manager
# Uncomment when using SSL
#    secrets:
#       - source: public_crt
#         target: /etc/nginx/cert/server.crt
#       - source: private_key
#         target: /etc/nginx/cert/server.key
    networks:
      - harbor
    ports:
      - 80:80
      - 443:443
      - 4443:4443

networks:
  harbor:
    driver: overlay

volumes:
  redis:
  jobs:
  core_ca_download:
  core_psc:
  admin_config:
  admin_data:
  database:
  registry:

secrets:
  secretkey:
    file: "./data/secretkey"
  registry_root_crt:
    file: "./common/config/registry/root.crt"
  core_privatekey:
    file: "./common/config/core/private_key.pem"
# Uncomment when using SSL
#  public_crt:
#    file: "./common/config/nginx/cert/server.crt"
#  private_key:
#    file: "./common/config/nginx/cert/server.key"

configs:
  registry_config:
    file: "./common/config/registry/config.yml"
  registryctl_config:
    file: "./common/config/registryctl/config.yml"
  jobservice_config:
    file: "./common/config/jobservice/config.yml"
  core_config:
    file: "./common/config/core/app.conf"
# Uncomment when using clair
#  clair_config:
#    file: "./common/config/clair/config.yaml"
  nginx_config:
    file: "./common/config/nginx/nginx.conf"
