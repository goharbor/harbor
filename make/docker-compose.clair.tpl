version: '2'
services:
  core:
    networks:
      harbor-clair:
        aliases:
          - harbor-core
  jobservice:
    networks:
      - harbor-clair
  registry:
    networks:
      - harbor-clair
  postgresql:
    networks:
      harbor-clair:
        aliases:
          - harbor-db
  clair:
    networks:
      - harbor-clair
    container_name: clair
    image: goharbor/clair-photon:__clair_version__
    restart: always
    cpu_quota: 50000
    dns_search: .
    depends_on:
      - postgresql
    volumes:
      - ./common/config/clair/config.yaml:/etc/clair/config.yaml:z
    logging:
      driver: "syslog"
      options:  
        syslog-address: "tcp://127.0.0.1:1514"
        tag: "clair"
    env_file:
      ./common/config/clair/clair_env
networks:
  harbor-clair:
    external: false
